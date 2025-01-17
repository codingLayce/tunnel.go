package tcp

import (
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer checks a nominal server case.
// - Starts correctly
// - Accept 2 connections
// - On connection is closed by the client
// - The other one stays open until server stop
// - Stops correctly
func TestServer(t *testing.T) {
	connectionReceived := make(chan struct{})
	connectionClosed := make(chan struct{})
	srv := NewServer(&ServerOption{
		Addr: ":19917",
		OnConnectionReceived: func(conn *Connection) {
			connectionReceived <- struct{}{}
		},
		OnConnectionClosed: func(conn *Connection, _ bool) {
			connectionClosed <- struct{}{}
		},
	})

	err := srv.Start()
	require.NoError(t, err)
	slog.Debug("Server started")

	// Initiate connection 1 and wait for connection callback
	conn1, err := net.Dial("tcp", ":19917")
	require.NoError(t, err)

	select {
	case <-connectionReceived:
		slog.Debug("Connection received")
	case <-time.After(50 * time.Millisecond):
		assert.FailNow(t, "A connection should have been received")
	}

	// Initiate connection 2 and wait for connection callback
	_, err = net.Dial("tcp", ":19917")
	require.NoError(t, err)

	select {
	case <-connectionReceived:
		slog.Debug("Connection received")
	case <-time.After(50 * time.Millisecond):
		assert.FailNow(t, "A connection should have been received")
	}

	// Close connection 1 and wait for connection closed callback
	err = conn1.Close()
	require.NoError(t, err)

	select {
	case <-connectionClosed:
		slog.Debug("Connection closed")
	case <-time.After(50 * time.Millisecond):
		assert.FailNow(t, "A connection should have been closed")
	}

	// Stop the server
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		srv.Stop()
	}()

	// Connection 2 should close and invoke the callback
	select {
	case <-connectionClosed:
		slog.Debug("Connection closed")
	case <-time.After(50 * time.Millisecond):
		assert.FailNow(t, "A connection should have been closed")
	}

	// Graceful shutdown wait
	select {
	case <-stopped:
		slog.Debug("Server stopped")
	case <-time.After(50 * time.Millisecond):
		assert.FailNow(t, "Server should have stopped")
	}
}
