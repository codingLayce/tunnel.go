package tunnel

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()
}

func TestConnect_ServerDisconnecting(t *testing.T) {
	// The TCPClient mock is waiting for the error to be returned by the connect method on call to it.
	connectCalled := make(chan error)
	tcpClient := newTestTCPClient()
	tcpClient.connect = func() error {
		tcpClient.done = make(chan struct{})
		return <-connectCalled // Blocks until it gets the error to return and returns it.
	}
	mockNewTCPClient(t, tcpClient)

	var (
		mtx sync.Mutex
		cl  *Client
	)
	go func() {
		c, err := Connect("")
		require.NoError(t, err)
		mtx.Lock()
		cl = c
		mtx.Unlock()
	}()

	// Wait for the first connect and make it complete
	select {
	case connectCalled <- nil:
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "Connect should have been called")
	}

	// Stop the internal connection
	close(tcpClient.done)

	// Wait for the automatic connection retry and makes it fail
	select {
	case connectCalled <- errors.New("error"):
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "Connect should have been called")
	}

	// Wait for the next connection try 1s later and make it complete
	select {
	case connectCalled <- nil:
	case <-time.After(2 * time.Second):
		assert.FailNow(t, "Connect should have been called")
	}

	// Manually close
	mtx.Lock()
	cl.Stop()
	mtx.Unlock()

	// No Connect should be called
	select {
	case connectCalled <- nil:
		assert.FailNow(t, "Connect shouldn't have been called")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestClient_StoppedWhileRetryConnect(t *testing.T) {
	// First time connect is called will work.
	tcpClient := newTestTCPClient()
	connectCalled := atomic.Int32{}
	tcpClient.connect = func() error {
		tcpClient.done = make(chan struct{})
		connectCalled.Add(1)
		if connectCalled.Load() == 1 {
			return nil
		}
		return errors.New("error")
	}
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	close(tcpClient.done)
	time.Sleep(100 * time.Millisecond)

	stopped := make(chan struct{})
	go func() {
		cl.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		assert.FailNow(t, "Client should have stopped")
	}
}

func TestClient_OnPayload(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	// Nothing to assert yet. The log should be in the standard output.
	assert.NotPanics(t, func() {
		cl, err := Connect("")
		require.NoError(t, err)
		defer cl.Stop()

		tcpClient.callOnPayload([]byte("PAYLOAD"))
	})
}

func TestConnect_Error(t *testing.T) {
	tcpClient := newTestTCPClient()
	tcpClient.connect = func() error {
		return errors.New("error")
	}
	mockNewTCPClient(t, tcpClient)

	_, err := Connect("")
	assert.EqualError(t, err, "connect to Tunnel server: error")
}
