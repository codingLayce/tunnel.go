package tunnel

import (
	"log/slog"
	"testing"

	"github.com/codingLayce/tunnel.go/tcp"
)

type TestTCPClient struct {
	connect   func() error
	done      chan struct{}
	onPayload func([]byte)
}

func newTestTCPClient() *TestTCPClient { return &TestTCPClient{done: make(chan struct{})} }

func (t *TestTCPClient) Connect() error {
	t.done = make(chan struct{})
	if t.connect != nil {
		return t.connect()
	}
	return nil
}
func (t *TestTCPClient) Stop()                 {}
func (t *TestTCPClient) Done() <-chan struct{} { return t.done }
func (t *TestTCPClient) Send(_ []byte) error   { return nil }

func (t *TestTCPClient) callOnPayload(payload []byte) {
	if t.onPayload != nil {
		t.onPayload(payload)
	}
}

func mockNewTCPClient(t *testing.T, client *TestTCPClient) {
	actualFn := newTCPClient
	newTCPClient = func(opts *tcp.ClientOption) TCPClient {
		client.onPayload = opts.OnPayload
		return client
	}
	t.Cleanup(func() {
		newTCPClient = actualFn
	})
}

func TestMain(m *testing.M) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	m.Run()
}
