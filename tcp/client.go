package tcp

import (
	"fmt"
	"net"
	"sync"
)

type ClientOption struct {
	Addr string

	// OnPayload is invoked when the server has sent a payload.
	OnPayload func(payload []byte)
}

type Client struct {
	opts *ClientOption

	conn *Connection

	wg      sync.WaitGroup
	stopped chan struct{}
}

func NewClient(opts *ClientOption) *Client {
	return &Client{
		opts:    opts,
		stopped: make(chan struct{}),
	}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.opts.Addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	c.conn = NewConnection(conn, &ConnectionOption{
		OnPayload: func(_ *Connection, payload []byte) {
			if c.opts.OnPayload != nil {
				c.opts.OnPayload(payload)
			}
		},
		OnConnectionClosed: func(_ *Connection, _ bool) {
			select { // prevent closing a closed chan
			case <-c.stopped:
			default:
				close(c.stopped)
			}
		},
	})

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		c.conn.payloadLoop()
	}()

	return nil
}

func (c *Client) Stop() {
	c.conn.Close()
	c.wg.Wait()
	select { // prevent closing a closed chan
	case <-c.stopped:
	default:
		close(c.stopped)
	}
}

func (c *Client) Done() <-chan struct{} {
	return c.stopped
}

func (c *Client) Send(payload []byte) error {
	return c.conn.Send(payload)
}
