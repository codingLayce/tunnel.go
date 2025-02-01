package tunnel

import (
	"log/slog"
	"testing"

	"github.com/codingLayce/tunnel.go/pdu"
	"github.com/codingLayce/tunnel.go/pdu/command"
	"github.com/codingLayce/tunnel.go/tcp"
	"github.com/codingLayce/tunnel.go/test-helper/mock"
)

type TestTCPClient struct {
	connect   func() error
	done      chan struct{}
	onPayload func([]byte)
	send      func([]byte) error
	cmdCh     chan command.Command
}

func newTestTCPClient() *TestTCPClient {
	return &TestTCPClient{done: make(chan struct{}), cmdCh: make(chan command.Command)}
}

func (t *TestTCPClient) Connect() error {
	t.done = make(chan struct{})
	if t.connect != nil {
		return t.connect()
	}
	return nil
}
func (t *TestTCPClient) Stop() {
	if t.cmdCh != nil {
		select {
		case <-t.cmdCh:
		default:
			close(t.cmdCh)
		}
	}
}
func (t *TestTCPClient) Done() <-chan struct{} { return t.done }
func (t *TestTCPClient) Send(payload []byte) error {
	if t.send != nil {
		return t.send(payload)
	}

	cmd, err := pdu.Unmarshal(payload)
	if err != nil {
		return err
	}

	select {
	case t.cmdCh <- cmd:
	case <-t.done:
	}

	return nil
}

func (t *TestTCPClient) callOnPayload(payload []byte) {
	if t.onPayload != nil {
		t.onPayload(payload)
	}
}

func (t *TestTCPClient) commandsChan() <-chan command.Command {
	return t.cmdCh
}

func mockNewTCPClient(t *testing.T, client *TestTCPClient) {
	mock.Do(t, &newTCPClient, func(opts *tcp.ClientOption) TCPClient {
		client.onPayload = opts.OnPayload
		return client
	})
}

func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelInfo)

	m.Run()
}
