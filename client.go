package tunnel

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/codingLayce/tunnel.go/tcp"
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
		addr:   addr,
		Logger: slog.Default().With("entity", "TUNNEL_CLIENT"),
		stop:   make(chan struct{}),
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

func (c *Client) Stop() {
	close(c.stop)
	c.wg.Wait()
}

func (c *Client) onPayload(payload []byte) {
	c.Logger.Debug("Received payload from server", "payload_string", string(payload))
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

// retryToConnect retries to connect to the configured Tunnel server.
// Returns true if it succeeds reconnect, false otherwise.
// The delay between retries is incrementing by 20%.
// After 30 tries, the delay is around 3m17s and the time spend retrying is around 16m24s.
func (c *Client) retryToConnect() bool {
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
