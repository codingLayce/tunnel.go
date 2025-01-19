package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListenTunnel_Validate(t *testing.T) {
	for name, tc := range map[string]struct {
		name string
	}{
		"Valid only lower cases letters": {
			name: "bidule",
		},
		"Valid only upper case letters": {
			name: "BONJOUR",
		},
		"Valid only letters": {
			name: "Tunnel",
		},
		"Valid only digits": {
			name: "17",
		},
		"Valid letters and digits": {
			name: "88My67Tunnel01",
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := NewListenTunnel(tc.name)
			assert.NoError(t, cmd.Validate())
		})
	}
}

func TestListenTunnel_Validate_Error(t *testing.T) {
	for name, tc := range map[string]struct {
		name string
	}{
		"Invalid empty": {
			name: "",
		},
		"Invalid spaces": {
			name: "My Tunnel",
		},
		"Invalid special characters": {
			name: "A_Supper-tunnel",
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := NewListenTunnel(tc.name)
			assert.EqualError(t, cmd.Validate(), "invalid name")
		})
	}
}

func TestListenTunnel_Info(t *testing.T) {
	for name, tc := range map[string]struct {
		name         string
		expectedInfo string
	}{
		"Valid only lower cases letters": {
			name:         "bidule",
			expectedInfo: "LISTEN_TUNNEL(bidule)",
		},
		"Valid only upper case letters": {
			name:         "BONJOUR",
			expectedInfo: "LISTEN_TUNNEL(BONJOUR)",
		},
		"Valid only letters": {
			name:         "Tunnel",
			expectedInfo: "LISTEN_TUNNEL(Tunnel)",
		},
		"Valid only digits": {
			name:         "17",
			expectedInfo: "LISTEN_TUNNEL(17)",
		},
		"Valid letters and digits": {
			name:         "88My67Tunnel01",
			expectedInfo: "LISTEN_TUNNEL(88My67Tunnel01)",
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := NewListenTunnel(tc.name)
			assert.Equal(t, tc.expectedInfo, cmd.Info())
		})
	}
}

func TestListenTunnel_ID(t *testing.T) {
	cmd := NewListenTunnel("Tunnel")
	assert.Equal(t, newIDValue, cmd.TransactionID())
}

func TestListenTunnel_Data(t *testing.T) {
	for name, tc := range map[string]struct {
		name         string
		tunnelType   TunnelType
		expectedData []byte
	}{
		"Valid only lower cases letters": {
			name:         "bidule",
			tunnelType:   BroadcastTunnel,
			expectedData: []byte("bidule"),
		},
		"Valid only upper case letters": {
			name:         "BONJOUR",
			tunnelType:   BroadcastTunnel,
			expectedData: []byte("BONJOUR"),
		},
		"Valid only letters": {
			name:         "Tunnel",
			tunnelType:   BroadcastTunnel,
			expectedData: []byte("Tunnel"),
		},
		"Valid only digits": {
			name:         "17",
			tunnelType:   BroadcastTunnel,
			expectedData: []byte("17"),
		},
		"Valid letters and digits": {
			name:         "88My67Tunnel01",
			tunnelType:   BroadcastTunnel,
			expectedData: []byte("88My67Tunnel01"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := NewListenTunnel(tc.name)
			assert.Equal(t, tc.expectedData, cmd.Data())
		})
	}
}

func TestListenTunnel_Indicator(t *testing.T) {
	cmd := NewListenTunnel("Tunnel")
	assert.Equal(t, ListenTunnelIndicator, cmd.Indicator())
}
