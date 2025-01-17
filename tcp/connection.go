package tcp

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/rs/xid"
)

type ConnectionOption struct {
	OnConnectionClosed func(conn *Connection, timeout bool)
	OnPayload          func(conn *Connection, payload []byte)
	ReadTimeout        time.Duration
}

func (opts *ConnectionOption) defaults() {
	if opts.ReadTimeout <= time.Second {
		opts.ReadTimeout = defaultReadTimeout
	}
}

type Connection struct {
	net.Conn

	ID   string
	opts *ConnectionOption
}

func NewConnection(conn net.Conn, opts *ConnectionOption) *Connection {
	opts.defaults()
	return &Connection{
		Conn: conn,
		ID:   xid.New().String(),
		opts: opts,
	}
}

func (c *Connection) Send(payload []byte) error {
	_, err := c.Write(payload)
	return err
}

func (c *Connection) payloadLoop() {
	reader := bufio.NewReader(c)
	for {
		err := c.SetReadDeadline(time.Now().Add(c.opts.ReadTimeout))
		if err != nil {
			c.handleReadError(fmt.Errorf("set read deadline: %w", err))
			return
		}

		payload, err := reader.ReadBytes('\n')
		switch {
		case err == nil:
			c.handlePayload(payload)
		default:
			c.handleReadError(err)
			return
		}
	}
}

func (c *Connection) handleReadError(err error) {
	timeout := false
	switch {
	case os.IsTimeout(err):
		c.Close()
		timeout = true
	}

	if c.opts.OnConnectionClosed != nil {
		c.opts.OnConnectionClosed(c, timeout)
	}
}

func (c *Connection) handlePayload(payload []byte) {
	if c.opts.OnPayload != nil {
		c.opts.OnPayload(c, payload)
	}
}
