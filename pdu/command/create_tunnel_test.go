package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTunnel_Info(t *testing.T) {
	assert.Equal(t, "CREATE_TUNNEL(Bidule)", NewCreateTunnel("Bidule").Info())
}

func TestCreateTunnel_TransactionID(t *testing.T) {
	assert.Equal(t, newIDValue, NewCreateTunnel("Bidule").TransactionID())
}

func TestCreateTunnel_Indicator(t *testing.T) {
	assert.Equal(t, CreateTunnelIndicator, NewCreateTunnel("Bidule").Indicator())
}

func TestCreateTunnel_Data(t *testing.T) {
	assert.Equal(t, data([]byte{0x00}, []byte("Bidule")), NewCreateTunnel("Bidule").Data())
}
