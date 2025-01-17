package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTunnel_Validate(t *testing.T) {
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
			cmd := NewCreateTunnel(tc.name)
			assert.NoError(t, cmd.Validate())
		})
	}
}

func TestCreateTunnel_Validate_Error(t *testing.T) {
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
			cmd := NewCreateTunnel(tc.name)
			assert.EqualError(t, cmd.Validate(), "invalid name")
		})
	}
}

func TestCreateTunnel_Info(t *testing.T) {
	for name, tc := range map[string]struct {
		name         string
		expectedInfo string
	}{
		"Valid only lower cases letters": {
			name:         "bidule",
			expectedInfo: "CREATE_TUNNEL(bidule)",
		},
		"Valid only upper case letters": {
			name:         "BONJOUR",
			expectedInfo: "CREATE_TUNNEL(BONJOUR)",
		},
		"Valid only letters": {
			name:         "Tunnel",
			expectedInfo: "CREATE_TUNNEL(Tunnel)",
		},
		"Valid only digits": {
			name:         "17",
			expectedInfo: "CREATE_TUNNEL(17)",
		},
		"Valid letters and digits": {
			name:         "88My67Tunnel01",
			expectedInfo: "CREATE_TUNNEL(88My67Tunnel01)",
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := NewCreateTunnel(tc.name)
			assert.Equal(t, tc.expectedInfo, cmd.Info())
		})
	}
}

func TestCreateTunnel_ID(t *testing.T) {
	cmd := NewCreateTunnel("Tunnel")
	assert.Equal(t, newIDValue, cmd.TransactionID())
}

func TestCreateTunnel_Data(t *testing.T) {
	for name, tc := range map[string]struct {
		name         string
		expectedData []byte
	}{
		"Valid only lower cases letters": {
			name:         "bidule",
			expectedData: []byte("bidule"),
		},
		"Valid only upper case letters": {
			name:         "BONJOUR",
			expectedData: []byte("BONJOUR"),
		},
		"Valid only letters": {
			name:         "Tunnel",
			expectedData: []byte("Tunnel"),
		},
		"Valid only digits": {
			name:         "17",
			expectedData: []byte("17"),
		},
		"Valid letters and digits": {
			name:         "88My67Tunnel01",
			expectedData: []byte("88My67Tunnel01"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := NewCreateTunnel(tc.name)
			assert.Equal(t, tc.expectedData, cmd.Data())
		})
	}
}

func TestCreateTunnel_Indicator(t *testing.T) {
	cmd := NewCreateTunnel("Tunnel")
	assert.Equal(t, CreateTunnelIndicator, cmd.Indicator())
}
