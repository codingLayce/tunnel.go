package tunnel

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
	"github.com/codingLayce/tunnel.go/tcp"
)

var (
	waitForAckTimeout = 10 * time.Second
)

type TCPClient interface {
	Connect() error
	Stop()
	Done() <-chan struct{}
	Send(payload []byte) error
}

type Client struct {
	addr     string
	internal TCPClient

	// ackWaiters stores channel used to wait for an acknowledgement of the transaction_id (key).
	// true is written when ack is received, false is written when nack is received.
	ackWaiters map[string]chan bool

	stop chan struct{}
	wg   sync.WaitGroup
	mtx  sync.Mutex

	Logger *slog.Logger
}

// Connect creates a new Client and connects to a Tunnel server.
// You must call Stop method when the client is no longer needed (to gracefully wait for the internal routines to finish).
//
// Internally the Client is going to keep the connection with the server active (by retrying to connect with the Tunnel server if the connection is lost).
func Connect(addr string) (*Client, error) {
	client := &Client{
		addr:       addr,
		Logger:     slog.Default().With("entity", "TUNNEL_CLIENT"),
		ackWaiters: make(map[string]chan bool),
		stop:       make(chan struct{}),
	}
	client.internal = newTCPClient(&tcp.ClientOption{
		Addr:      addr,
		OnPayload: client.onPayload,
	})

	err := client.internal.Connect()
	if err != nil {
		return nil, fmt.Errorf("connect to Tunnel server: %w", err)
	}

	client.Logger.Debug("Connected to Tunnel server")

	client.wg.Add(1)
	go client.keepConnectedLoop()

	return client, nil
}

// Stop stops the internal client.
// Client is no longer usable after stopping it.
func (c *Client) Stop() {
	close(c.stop)
	c.wg.Wait()
}

// CreateBTunnel asks the server to create a new Broadcast Tunnel.
// Returns an error if the name is invalid or if the server nack the request.
func (c *Client) CreateBTunnel(name string) error {
	cmd := command.NewCreateTunnel(name)
	cmd.Type = command.BroadcastTunnel
	err := cmd.Validate()
	if err != nil {
		return fmt.Errorf("validate command: %w", err)
	}

	payload := pdu.Marshal(cmd)
	err = c.internal.Send(payload)
	if err != nil {
		return fmt.Errorf("send command: %w", err)
	}

	c.Logger.Debug("Sent command to server", "transaction_id", cmd.TransactionID(), "command", cmd.Info())

	ackCh := c.newAckWaiter(cmd.TransactionID())
	defer c.unstoreAckWaiter(cmd.TransactionID())

	select {
	case ack := <-ackCh:
		if ack {
			c.Logger.Info("Tunnel created", "tunnel", cmd.Info())
			return nil
		}
		c.Logger.Warn("Tunnel not created", "tunnel", cmd.Info())
		return fmt.Errorf("server refuses to create Tunnel")
	case <-time.After(waitForAckTimeout):
		c.Logger.Error("Timeout waiting for server acknowledgement")
		return fmt.Errorf("timeout waiting for server acknowledgement")
	}
}

func (c *Client) onPayload(payload []byte) {
	// Invoked by the tcpConnection when a payload is received.
	// IT'S BLOCKING THE TCP CONNECTION.

	c.Logger.Debug("Received payload from server", "payload_string", string(payload))

	cmd, err := pdu.Unmarshal(payload)
	if err != nil {
		c.Logger.Warn("Received unparsable payload. Discarding it.", "error", err)
		return
	}

	switch cmd.(type) {
	case *command.Ack:
		c.acknowledgementReceived(cmd.TransactionID(), true)
	case *command.Nack:
		c.acknowledgementReceived(cmd.TransactionID(), false)
	default:
		c.Logger.Warn("Received unsupported command", "command", cmd.Info())
	}
}

func (c *Client) acknowledgementReceived(transactionID string, isAck bool) {
	c.mtx.Lock()
	waiter, ok := c.ackWaiters[transactionID]
	c.mtx.Unlock()
	if !ok {
		c.Logger.Warn("Received unexpected acknowledgement. Discarding it.")
		return
	}
	waiter <- isAck
}

func (c *Client) newAckWaiter(transactionID string) <-chan bool {
	// Creates a channel and stores it into ackWaiters.

	ch := make(chan bool)
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.ackWaiters[transactionID] = ch
	return ch
}

func (c *Client) unstoreAckWaiter(transactionID string) {
	// Closes and delete the channel from ackWaiters.
	c.mtx.Lock()
	defer c.mtx.Unlock()
	ch, ok := c.ackWaiters[transactionID]
	if ok {
		close(ch)
		delete(c.ackWaiters, transactionID)
	}
}

func (c *Client) keepConnectedLoop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.stop:
			c.internal.Stop()
			c.Logger.Debug("Client asked to stop")
			return
		case <-c.internal.Done():
			c.Logger.Debug("Connection lost with Tunnel server. Reconnecting...")
			c.resetInternal()
			hasReconnect := c.retryToConnect()
			if !hasReconnect {
				c.Logger.Debug("Client asked to stop. Not reconnected")
				return
			}
			c.Logger.Debug("Reconnected to Tunnel server !")
		}
	}
}

func (c *Client) retryToConnect() bool {
	// retries to connect to the configured Tunnel server.
	// Returns true if it succeeds reconnect, false otherwise.
	// The delay between retries is incrementing by 20%.
	// After 30 tries, the delay is around 3m17s and the time spend retrying is around 16m24s.
	// TODO: Implement a max retries
	delay := time.Second
	err := c.internal.Connect()
	for err != nil {
		select {
		case <-c.stop:
			return false
		case <-time.After(delay):
			delay = time.Duration(float64(delay) * 1.2)
			err = c.internal.Connect()
		}
	}
	return true
}

func (c *Client) resetInternal() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.internal = newTCPClient(&tcp.ClientOption{
		Addr:      c.addr,
		OnPayload: c.onPayload,
	})
}

var newTCPClient = func(opts *tcp.ClientOption) TCPClient {
	return tcp.NewClient(opts)
}

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}
