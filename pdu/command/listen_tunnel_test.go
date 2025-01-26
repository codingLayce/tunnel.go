package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListenTunnel_Info(t *testing.T) {
	assert.Equal(t, "LISTEN_TUNNEL(Bidule)", NewListenTunnel("Bidule").Info())
}

func TestListenTunnel_TransactionID(t *testing.T) {
	assert.Equal(t, newIDValue, NewListenTunnel("Bidule").TransactionID())
}

func TestListenTunnel_Indicator(t *testing.T) {
	assert.Equal(t, ListenTunnelIndicator, NewListenTunnel("Bidule").Indicator())
}

func TestListenTunnel_Data(t *testing.T) {
	assert.Equal(t, []byte("Bidule"), NewListenTunnel("Bidule").Data())
}
