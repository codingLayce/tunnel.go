package tunnel

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codingLayce/tunnel/tcp"
)

type TCPClient interface {
	Connect() error
	Stop()
	Done() <-chan struct{}
	Send(payload []byte) error
}

type Client struct {
	internal  TCPClient
	connected atomic.Bool

	stop chan struct{}
	wg   sync.WaitGroup

	Logger *slog.Logger
}

func (c *Client) onPayload(payload []byte) {
	c.Logger.Debug("Received payload from server", "payload_string", string(payload))
}

func (c *Client) keepConnectedLoop() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.stop:
				c.internal.Stop()
				c.connected.Store(false)
				c.Logger.Debug("Stopped")
				return
			case <-c.internal.Done():
				c.connected.Store(false)
				c.Logger.Debug("Connection lost with Tunnel server. Reconnecting...")
				c.internal = newTCPClient(&tcp.ClientOption{
					// TODO : Allow to dial on a custom address
					Addr:      ":19917",
					OnPayload: c.onPayload,
				})

				err := c.internal.Connect()
				for err != nil {
					// TODO : Backoff retry or something less brutal !
					time.Sleep(time.Second)
					select {
					case <-c.stop:
						c.connected.Store(false)
						c.Logger.Debug("Stopped")
						return
					default:
					}
					err = c.internal.Connect()
				}
				c.Logger.Debug("Reconnected to Tunnel server !")
				c.connected.Store(true)
			}
		}
	}()
}

func (c *Client) Stop() {
	close(c.stop)
	c.wg.Wait()
}

// Connect creates a new Client and connects to a Tunnel server.
// You must call Stop method when the client is no longer needed (for a graceful shutdown).
//
// Internally the Client is going to keep the connection with the server active (by retrying to connect with the Tunnel server if the connection is lost).
func Connect() (*Client, error) {
	client := &Client{
		Logger: slog.Default().With("entity", "TUNNEL_CLIENT"),
		stop:   make(chan struct{}),
	}
	client.internal = newTCPClient(&tcp.ClientOption{
		// TODO : Allow to dial on a custom address
		Addr:      ":19917",
		OnPayload: client.onPayload,
	})

	err := client.internal.Connect()
	if err != nil {
		return nil, fmt.Errorf("connect to Tunnel server: %w", err)
	}
	client.connected.Store(true)
	client.Logger.Debug("Connected to Tunnel server")

	client.keepConnectedLoop()

	return client, nil
}

var newTCPClient = func(opts *tcp.ClientOption) TCPClient {
	return tcp.NewClient(opts)
}

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}
