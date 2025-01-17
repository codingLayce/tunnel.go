package tcp

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	serverSideClient := make(chan *Connection)
	server := NewServer(&ServerOption{
		Addr: ":19917",
		OnConnectionReceived: func(conn *Connection) {
			serverSideClient <- conn
		},
	})

	err := server.Start()
	require.NoError(t, err)
	defer server.Stop()
	slog.Debug("Server started")

	payloadReceived := make(chan []byte)
	cl := NewClient(&ClientOption{
		Addr: ":19917",
		OnPayload: func(payload []byte) {
			payloadReceived <- payload
		},
	})

	err = cl.Connect()
	require.NoError(t, err)
	slog.Debug("Client connected")

	select {
	case conn := <-serverSideClient:
		go func() {
			err = conn.Send([]byte("Bonjour\n"))
			require.NoError(t, err)
			slog.Debug("Server sent payload")
		}()
	case <-time.After(50 * time.Millisecond):
		assert.FailNow(t, "A connection should have been initiated on the server")
	}

	select {
	case payload := <-payloadReceived:
		assert.Equal(t, []byte("Bonjour\n"), payload)
		slog.Debug("Client received payload")
	case <-time.After(50 * time.Millisecond):
		assert.FailNow(t, "A payload should have been received to the client")
	}

	// Stop the client
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		cl.Stop()
	}()

	select {
	case <-stopped:
		slog.Debug("Client stopped")
	case <-time.After(50 * time.Millisecond):
		assert.FailNow(t, "Client should have stopped")
	}
}
