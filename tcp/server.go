package tcp

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const defaultReadTimeout = time.Minute

type ServerOption struct {
	Addr string

	// OnConnectionReceived is invoked when a connection is accepted by the server.
	// It's invoked inside the connection goroutine so it doesn't block server.
	OnConnectionReceived func(conn *Connection)

	// OnConnectionClosed is invoked when a connection is closed by the server or the client.
	// It's invoked inside the connection goroutine so it doesn't block server.
	OnConnectionClosed func(conn *Connection, timeout bool)

	// OnPayload is invoked when a connection has sent a payload.
	// It's invoked inside the connection goroutine and blocks next read.
	OnPayload func(conn *Connection, payload []byte)

	// ReadTimeout is the allowed idle duration before disconnecting the client.
	ReadTimeout time.Duration
}

func (opts *ServerOption) defaults() {
	if opts.ReadTimeout <= time.Second {
		opts.ReadTimeout = defaultReadTimeout
	}
}

type Server struct {
	opts *ServerOption

	connections map[string]*Connection
	listener    net.Listener

	stopped chan struct{}
	mtx     sync.Mutex
	wg      sync.WaitGroup
}

func NewServer(opts *ServerOption) *Server {
	opts.defaults()
	return &Server{
		opts:        opts,
		connections: make(map[string]*Connection),
		stopped:     make(chan struct{}),
	}
}

func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.opts.Addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	go s.acceptLoop()

	return nil
}

func (s *Server) acceptLoop() {
	s.wg.Add(1)
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		switch {
		case err == nil:
			s.handleConnection(conn)
		case errors.Is(err, net.ErrClosed):
			return
		default:
			s.Stop()
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	connection := NewConnection(conn, &ConnectionOption{
		OnConnectionClosed: s.opts.OnConnectionClosed,
		OnPayload:          s.opts.OnPayload,
		ReadTimeout:        s.opts.ReadTimeout,
	})
	s.storeConnection(connection)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.deleteConnection(connection.ID)

		if s.opts.OnConnectionReceived != nil {
			s.opts.OnConnectionReceived(connection)
		}

		connection.payloadLoop()
	}()
}

func (s *Server) storeConnection(connection *Connection) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.connections[connection.ID] = connection
}

func (s *Server) deleteConnection(id string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.connections, id)
}

func (s *Server) stopConnections() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	for _, connection := range s.connections {
		connection.Close()
	}
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) Stop() {
	s.stopConnections()
	s.listener.Close()
	s.wg.Wait()
	select { // prevent closing a closed chan
	case <-s.stopped:
	default:
		close(s.stopped)
	}
}

func (s *Server) Done() <-chan struct{} {
	return s.stopped
}
