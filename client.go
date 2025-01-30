package tunnel

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/codingLayce/tunnel.go/common/maps"
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
	ackWaiters *maps.SyncMap[string, chan bool]

	// listeners stores channel used to receive message from the given tunnel name (key).
	// The raw string message is passed to the channel.
	// /!\ Currently there is no way to stop listening /!\
	listeners *maps.SyncMap[string, chan string]

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
	// TODO: Configurations for logger

	client := &Client{
		addr:       addr,
		Logger:     slog.Default().With("entity", "TUNNEL_CLIENT"),
		ackWaiters: maps.NewSyncMap[string, chan bool](),
		listeners:  maps.NewSyncMap[string, chan string](),
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

	client.Logger.Info("Connected to Tunnel server")

	client.wg.Add(1)
	go client.keepConnectedLoop()

	return client, nil
}

// Stop stops the internal client.
// Client is no longer usable after stopping it.
func (c *Client) Stop() {
	close(c.stop)
	c.wg.Wait()
	c.Logger.Info("Stopped")
}

// PublishMessage publishes the given message to the given Tunnel.
// Returns an error if the server doesn't accept the message.
func (c *Client) PublishMessage(tunnelName, message string) error {
	cmd := command.NewPublishMessage(tunnelName, message)
	err := c.sendCommand(cmd)
	if err != nil {
		return err
	}

	err = c.waitAck(cmd.TransactionID())
	if err != nil {
		// TODO: Better error handling (typed error returned to client)
		c.Logger.Error("Error waiting for ack", "error", err)
		return err
	}

	return nil
}

// ListenTunnel makes the client listening for the given Tunnel's name messages.
// When a message is received, the callback function is invoked.
func (c *Client) ListenTunnel(name string, callback func(string)) error {
	cmd := command.NewListenTunnel(name)
	err := c.sendCommand(cmd)
	if err != nil {
		// TODO: Better error handling (typed error returned to client)
		return err
	}

	err = c.waitAck(cmd.TransactionID())
	if err != nil {
		// TODO: Better error handling (typed error returned to client)
		c.Logger.Error("Error waiting for ack", "error", err)
		return err
	}

	c.wg.Add(1)
	go c.listenTunnel(name, callback)

	return nil
}

// CreateBTunnel asks the server to create a new Broadcast Tunnel.
// Returns an error if the name is invalid or if the server nack the request.
func (c *Client) CreateBTunnel(name string) error {
	cmd := command.NewCreateTunnel(name)
	cmd.Type = command.BroadcastTunnel

	err := c.sendCommand(cmd)
	if err != nil {
		// TODO: Better error handling (typed error returned to client)
		return err
	}

	err = c.waitAck(cmd.TransactionID())
	if err != nil {
		// TODO: Better error handling (typed error returned to client)
		c.Logger.Error("Error waiting for ack", "error", err)
		return err
	}

	return nil
}

func (c *Client) listenTunnel(tunnelName string, callback func(string)) {
	defer c.wg.Done()
	msgCh := make(chan string)
	c.listeners.Put(tunnelName, msgCh)

	for {
		select {
		case msg := <-msgCh:
			c.Logger.Debug("Received message", "tunnel_name", tunnelName, "message", msg)
			callback(msg)
		case <-c.stop:
			c.Logger.Debug("Stop listening Tunnel", "tunnel_name", tunnelName)
			return
		}
	}
}

func (c *Client) waitAck(transactionID string) error {
	ackCh := make(chan bool)
	c.ackWaiters.Put(transactionID, ackCh)
	defer c.unstoreAckWaiter(transactionID)

	select {
	case ack := <-ackCh:
		if ack {
			return nil
		}
		return fmt.Errorf("server nack")
	case <-time.After(waitForAckTimeout):
		return fmt.Errorf("timeout waiting for server acknowledgement")
	}
}

func (c *Client) sendCommand(cmd command.Command) error {
	err := cmd.Validate()
	if err != nil {
		return fmt.Errorf("validate command: %w", err)
	}

	payload := pdu.Marshal(cmd)
	c.Logger.Debug("Sending payload", "payload", payload)

	err = c.internal.Send(payload)
	if err != nil {
		return fmt.Errorf("send command: %w", err)
	}

	c.Logger.Info("Sent command", "transaction_id", cmd.TransactionID(), "command", cmd.Info())
	return nil
}

func (c *Client) onPayload(payload []byte) {
	// Invoked by the tcpConnection when a payload is received.
	// IT'S BLOCKING THE TCP CONNECTION.

	c.Logger.Debug("Received payload", "payload", payload)

	cmd, err := pdu.Unmarshal(payload)
	if err != nil {
		c.Logger.Warn("Received unparsable payload. Discarding it.", "error", err)
		return
	}

	c.Logger.Info("Received command", "transaction_id", cmd.TransactionID(), "command", cmd.Info())

	switch castedCMD := cmd.(type) {
	case *command.Ack:
		c.acknowledgementReceived(cmd.TransactionID(), true)
	case *command.Nack:
		c.acknowledgementReceived(cmd.TransactionID(), false)
	case *command.ReceiveMessage:
		c.messageReceived(castedCMD)
	default:
		c.Logger.Warn("Received unsupported command", "command", cmd.Info())
	}
}

func (c *Client) acknowledgementReceived(transactionID string, isAck bool) {
	waiter, ok := c.ackWaiters.Get(transactionID)
	if !ok {
		c.Logger.Warn("Received unexpected acknowledgement. Discarding it.")
		return
	}
	// TODO: Could be blocking if nobody listen to it (shouldn't happen but you know..)
	waiter <- isAck
}

func (c *Client) messageReceived(cmd *command.ReceiveMessage) {
	ch, ok := c.listeners.Get(cmd.TunnelName)
	if !ok {
		c.Logger.Error("No listener for the received message", "tunnel_name", cmd.TunnelName)
		return
	}
	select { // Prevent blocking when the Client is stopped.
	case ch <- cmd.Message:
	case <-c.stop:
	}
}

func (c *Client) unstoreAckWaiter(transactionID string) {
	// Closes and delete the channel from ackWaiters.
	ch, ok := c.ackWaiters.Get(transactionID)
	if ok {
		close(ch)
		c.ackWaiters.Delete(transactionID)
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
			c.Logger.Debug("Connection lost with Tunnel server")
			c.resetInternal()
			hasReconnect := c.retryToConnect()
			if !hasReconnect {
				c.Logger.Debug("Client asked to stop. Not reconnected")
				return
			}
			c.Logger.Debug("Reconnected !")
		}
	}
}

func (c *Client) retryToConnect() bool {
	// retries to connect to the configured Tunnel server.
	// Returns true if it succeeds reconnect, false otherwise.
	// The delay between retries is incrementing by 20%.
	// After 30 tries, the delay is around 3m17s and the time spend retrying is around 16m24s.
	// TODO: Implement a max retries
	// TODO: Introduce retry policy to allow the user to choose
	delay := time.Second
	c.Logger.Debug("Retry to connect...")
	err := c.internal.Connect()
	for err != nil {
		c.Logger.Debug("Cannot reach Tunnel server. Retrying after delay", "delay", delay)
		select {
		case <-c.stop:
			return false
		case <-time.After(delay):
			delay = time.Duration(float64(delay) * 1.2).Round(time.Millisecond)
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
