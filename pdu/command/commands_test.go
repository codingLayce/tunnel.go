package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	transactionID := "abcd1234"
	for name, tc := range map[string]struct {
		indicator       byte
		data            []byte
		expectedCommand Command
	}{
		"Ack": {
			indicator:       AcknowledgementIndicator,
			data:            []byte(ackData),
			expectedCommand: NewAckWithTransactionID(transactionID),
		},
		"Nack": {
			indicator:       AcknowledgementIndicator,
			data:            []byte(nackData),
			expectedCommand: NewNackWithTransactionID(transactionID),
		},
		"Create Tunnel": {
			indicator:       CreateTunnelIndicator,
			data:            data([]byte{0}, []byte("Bidule17")),
			expectedCommand: NewCreateTunnelWithTransactionID(transactionID, "Bidule17"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd, err := Parse(tc.indicator, transactionID, tc.data)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCommand, cmd)
		})
	}
}

func TestParse_Errors(t *testing.T) {
	for name, tc := range map[string]struct {
		indicator        byte
		data             []byte
		expectedErrorMsg string
	}{
		"Unknown indicator": {
			indicator:        0xff,
			expectedErrorMsg: "invalid command indicator: unknown 0xff",
		},
		"Invalid acknowledgement": {
			indicator:        AcknowledgementIndicator,
			data:             []byte("Bidule"),
			expectedErrorMsg: `invalid acknowledgement command: unknown data "Bidule"`,
		},
		"Invalid create tunnel validation": {
			indicator:        CreateTunnelIndicator,
			data:             []byte("Mon_Tunnel"),
			expectedErrorMsg: `invalid create_tunnel command: invalid type`,
		},
		"Invalid create tunnel data": {
			indicator:        CreateTunnelIndicator,
			data:             []byte{},
			expectedErrorMsg: `invalid payload: missing tunnel type`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd, err := Parse(tc.indicator, "abcd1234", tc.data)
			assert.EqualError(t, err, tc.expectedErrorMsg)
			assert.Nil(t, cmd)
		})
	}

}
