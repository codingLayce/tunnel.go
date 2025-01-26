package tunnel

import (
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel.go/id"
	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
	"github.com/codingLayce/tunnel.go/test-helper/mock"
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
	case <-time.After(50 * time.Millisecond):
		assert.FailNow(t, "Connect should have been called")
	}

	// Stop the internal connection
	close(tcpClient.done)

	// Wait for the automatic connection retry and makes it fail
	select {
	case connectCalled <- errors.New("error"):
	case <-time.After(50 * time.Millisecond):
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
	case <-time.After(50 * time.Millisecond):
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

	time.Sleep(50 * time.Millisecond)
	close(tcpClient.done)
	time.Sleep(50 * time.Millisecond)

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

func TestConnect_Error(t *testing.T) {
	tcpClient := newTestTCPClient()
	tcpClient.connect = func() error {
		return errors.New("error")
	}
	mockNewTCPClient(t, tcpClient)

	_, err := Connect("")
	assert.EqualError(t, err, "connect to Tunnel server: error")
}

func TestClient_CreateBTunnel(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	go func() {
		select {
		case cmd := <-tcpClient.commandsChan():
			slog.Debug("[SERVER] Received command", "command", cmd.Info())
			createTunnel, ok := cmd.(*command.CreateTunnel)
			require.True(t, ok)
			assert.Equal(t, "MyTunnel", createTunnel.Name)
			// send ack
			time.Sleep(50 * time.Millisecond) // Let time to waiter to be created
			tcpClient.callOnPayload(pdu.Marshal(command.NewAckWithTransactionID(cmd.TransactionID())))
		case <-time.After(50 * time.Millisecond):
			assert.FailNow(t, "Server should have received a CreateTunnel command")
		}
	}()

	err = cl.CreateBTunnel("MyTunnel")
	require.NoError(t, err)
}

func TestClient_CreateBTunnel_ValidationError(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	err = cl.CreateBTunnel("Un super tunnel avec un mauvais _nom")
	assert.EqualError(t, err, "validate command: invalid name")
}

func TestClient_CreateBTunnel_SendError(t *testing.T) {
	tcpClient := newTestTCPClient()
	tcpClient.send = func(_ []byte) error {
		return errors.New("error")
	}
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	err = cl.CreateBTunnel("MonTunnel")
	assert.EqualError(t, err, "send command: error")
}

func TestClient_CreateBTunnel_NackError(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	go func() {
		select {
		case cmd := <-tcpClient.commandsChan():
			slog.Debug("[SERVER] Received command", "command", cmd.Info())
			createTunnel, ok := cmd.(*command.CreateTunnel)
			require.True(t, ok)
			assert.Equal(t, "MyTunnel", createTunnel.Name)
			// send nack
			time.Sleep(50 * time.Millisecond) // Let time to waiter to be created
			tcpClient.callOnPayload(pdu.Marshal(command.NewNackWithTransactionID(cmd.TransactionID())))
		case <-time.After(50 * time.Millisecond):
			assert.FailNow(t, "Server should have received a CreateTunnel command")
		}
	}()

	err = cl.CreateBTunnel("MyTunnel")
	assert.EqualError(t, err, "server nack")
}

func TestClient_CreateBTunnel_TimeoutError_NoResponse(t *testing.T) {
	mock.Do(t, &waitForAckTimeout, time.Second) // Not too long for tests execution

	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	go func() {
		select {
		case cmd := <-tcpClient.commandsChan():
			slog.Debug("[SERVER] Received command", "command", cmd.Info())
			createTunnel, ok := cmd.(*command.CreateTunnel)
			require.True(t, ok)
			assert.Equal(t, "MyTunnel", createTunnel.Name)
			// Not responding
		case <-time.After(50 * time.Millisecond):
			assert.FailNow(t, "Server should have received a CreateTunnel command")
		}
	}()

	err = cl.CreateBTunnel("MyTunnel")
	assert.EqualError(t, err, "timeout waiting for server acknowledgement")
}

func TestClient_CreateBTunnel_TimeoutError_WrongAckTransactionID(t *testing.T) {
	mock.Do(t, &waitForAckTimeout, time.Second) // Not too long for tests execution

	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	go func() {
		select {
		case cmd := <-tcpClient.commandsChan():
			slog.Debug("[SERVER] Received command", "command", cmd.Info())
			createTunnel, ok := cmd.(*command.CreateTunnel)
			require.True(t, ok)
			assert.Equal(t, "MyTunnel", createTunnel.Name)
			time.Sleep(50 * time.Millisecond) // Let time to waiter to be created
			// Respond with other transactionID
			tcpClient.callOnPayload(pdu.Marshal(command.NewAckWithTransactionID(id.New())))
		case <-time.After(50 * time.Millisecond):
			assert.FailNow(t, "Server should have received a CreateTunnel command")
		}
	}()

	err = cl.CreateBTunnel("MyTunnel")
	assert.EqualError(t, err, "timeout waiting for server acknowledgement")
}

func TestClient_CreateBTunnel_TimeoutError_InvalidPayload(t *testing.T) {
	mock.Do(t, &waitForAckTimeout, time.Second) // Not too long for tests execution

	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	go func() {
		select {
		case cmd := <-tcpClient.commandsChan():
			slog.Debug("[SERVER] Received command", "command", cmd.Info())
			createTunnel, ok := cmd.(*command.CreateTunnel)
			require.True(t, ok)
			assert.Equal(t, "MyTunnel", createTunnel.Name)
			// Sending unparsable payload
			tcpClient.callOnPayload([]byte{})
		case <-time.After(50 * time.Millisecond):
			assert.FailNow(t, "Server should have received a CreateTunnel command")
		}
	}()

	err = cl.CreateBTunnel("MyTunnel")
	assert.EqualError(t, err, "timeout waiting for server acknowledgement")
}

func TestClient_CreateBTunnel_TimeoutError_UnsupportedCommand(t *testing.T) {
	mock.Do(t, &waitForAckTimeout, time.Second) // Not too long for tests execution

	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	go func() {
		select {
		case cmd := <-tcpClient.commandsChan():
			slog.Debug("[SERVER] Received command", "command", cmd.Info())
			createTunnel, ok := cmd.(*command.CreateTunnel)
			require.True(t, ok)
			assert.Equal(t, "MyTunnel", createTunnel.Name)
			// Sending back the CreateTunnel commands (which doesn't make sense for a client to receive it)
			tcpClient.callOnPayload(pdu.Marshal(cmd))
		case <-time.After(50 * time.Millisecond):
			assert.FailNow(t, "Server should have received a CreateTunnel command")
		}
	}()

	err = cl.CreateBTunnel("MyTunnel")
	assert.EqualError(t, err, "timeout waiting for server acknowledgement")
}

func TestClient_ListenTunnel(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	go func() {
		select {
		case cmd := <-tcpClient.commandsChan():
			slog.Debug("[SERVER] Received command", "command", cmd.Info())
			listenTunnel, ok := cmd.(*command.ListenTunnel)
			require.True(t, ok)
			assert.Equal(t, "Bidule", listenTunnel.Name)
			// send ack
			time.Sleep(50 * time.Millisecond) // Let time to waiter to be created
			tcpClient.callOnPayload(pdu.Marshal(command.NewAckWithTransactionID(cmd.TransactionID())))

			time.Sleep(50 * time.Millisecond) // Let time to listener to be created
			// send message
			tcpClient.callOnPayload(pdu.Marshal(command.NewReceiveMessage(listenTunnel.Name, "This is a message")))
		case <-time.After(50 * time.Millisecond):
			assert.FailNow(t, "Server should have received a CreateTunnel command")
		}
	}()

	receivedMsg := make(chan string)
	err = cl.ListenTunnel("Bidule", func(msg string) {
		receivedMsg <- msg
	})
	require.NoError(t, err)

	select {
	case msg := <-receivedMsg:
		assert.Equal(t, "This is a message", msg)
	case <-time.After(100 * time.Millisecond):
		assert.FailNow(t, "A message should have been received")
	}
}

func TestClient_ListenTunnel_ValidationError(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	err = cl.ListenTunnel("Un super tunnel avec un mauvais _nom", func(_ string) {})
	assert.EqualError(t, err, "validate command: invalid name")
}

func TestClient_ListenTunnel_SendError(t *testing.T) {
	tcpClient := newTestTCPClient()
	tcpClient.send = func(_ []byte) error {
		return errors.New("error")
	}
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	err = cl.ListenTunnel("MonTunnel", func(_ string) {})
	assert.EqualError(t, err, "send command: error")
}

func TestClient_ListenTunnel_NackError(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	go func() {
		select {
		case cmd := <-tcpClient.commandsChan():
			slog.Debug("[SERVER] Received command", "command", cmd.Info())
			listenTunnel, ok := cmd.(*command.ListenTunnel)
			require.True(t, ok)
			assert.Equal(t, "MyTunnel", listenTunnel.Name)
			// send nack
			time.Sleep(50 * time.Millisecond) // Let time to waiter to be created
			tcpClient.callOnPayload(pdu.Marshal(command.NewNackWithTransactionID(cmd.TransactionID())))
		case <-time.After(50 * time.Millisecond):
			assert.FailNow(t, "Server should have received a ListenTunnel command")
		}
	}()

	err = cl.ListenTunnel("MyTunnel", func(_ string) {})
	assert.EqualError(t, err, "server nack")
}

func TestClient_PublishMessage(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	go func() {
		select {
		case cmd := <-tcpClient.commandsChan():
			slog.Debug("[SERVER] Received command", "command", cmd.Info())
			publishMessage, ok := cmd.(*command.PublishMessage)
			require.True(t, ok)
			assert.Equal(t, "Bidule", publishMessage.TunnelName)
			assert.Equal(t, "Mon super message", publishMessage.Message)
			// send ack
			time.Sleep(50 * time.Millisecond) // Let time to waiter to be created
			tcpClient.callOnPayload(pdu.Marshal(command.NewAckWithTransactionID(cmd.TransactionID())))
		case <-time.After(50 * time.Millisecond):
			assert.FailNow(t, "Server should have received a CreateTunnel command")
		}
	}()

	err = cl.PublishMessage("Bidule", "Mon super message")
	require.NoError(t, err)
}

func TestClient_PublishMessage_ValidationError(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	err = cl.PublishMessage("Un super tunnel avec un mauvais _nom", "Mon message")
	assert.EqualError(t, err, "validate command: invalid tunnel_name")
}

func TestClient_PublishMessage_SendError(t *testing.T) {
	tcpClient := newTestTCPClient()
	tcpClient.send = func(_ []byte) error {
		return errors.New("error")
	}
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	err = cl.PublishMessage("MonTunnel", "Mon message")
	assert.EqualError(t, err, "send command: error")
}

func TestClient_PublishMessage_NackError(t *testing.T) {
	tcpClient := newTestTCPClient()
	mockNewTCPClient(t, tcpClient)

	cl, err := Connect("")
	require.NoError(t, err)
	defer cl.Stop()

	go func() {
		select {
		case cmd := <-tcpClient.commandsChan():
			slog.Debug("[SERVER] Received command", "command", cmd.Info())
			publishMessage, ok := cmd.(*command.PublishMessage)
			require.True(t, ok)
			assert.Equal(t, "MyTunnel", publishMessage.TunnelName)
			assert.Equal(t, "Mon message", publishMessage.Message)
			// send nack
			time.Sleep(50 * time.Millisecond) // Let time to waiter to be created
			tcpClient.callOnPayload(pdu.Marshal(command.NewNackWithTransactionID(cmd.TransactionID())))
		case <-time.After(50 * time.Millisecond):
			assert.FailNow(t, "Server should have received a ListenTunnel command")
		}
	}()

	err = cl.PublishMessage("MyTunnel", "Mon message")
	assert.EqualError(t, err, "server nack")
}
