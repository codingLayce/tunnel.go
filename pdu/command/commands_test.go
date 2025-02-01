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
		"Listen Tunnel": {
			indicator:       ListenTunnelIndicator,
			data:            []byte("Bidule"),
			expectedCommand: NewListenTunnelWithTransactionID(transactionID, "Bidule"),
		},
		"Publish Message": {
			indicator:       PublishMessageIndicator,
			data:            data([]byte("TunnelName"), []byte{' '}, []byte("Mon super message")),
			expectedCommand: NewPublishMessageWithTransactionID(transactionID, "TunnelName", "Mon super message"),
		},
		"Receive Message": {
			indicator:       ReceiveMessageIndicator,
			data:            data([]byte("TunnelName"), []byte{' '}, []byte("Mon super message")),
			expectedCommand: NewReceiveMessageWithTransactionID(transactionID, "TunnelName", "Mon super message"),
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
		"Acknowledgement invalid payload": {
			indicator:        AcknowledgementIndicator,
			data:             []byte("Bidule"),
			expectedErrorMsg: `invalid acknowledgement command: unknown data "Bidule"`,
		},
		"Create_tunnel invalid payload": {
			indicator:        CreateTunnelIndicator,
			data:             []byte{},
			expectedErrorMsg: `invalid payload: missing tunnel type`,
		},
		"Create_tunnel invalid validation - Type": {
			indicator:        CreateTunnelIndicator,
			data:             []byte("Mon_Tunnel"),
			expectedErrorMsg: `invalid create_tunnel command: invalid type`,
		},
		"Create_tunnel invalid validation - Tunnel Name": {
			indicator:        CreateTunnelIndicator,
			data:             data([]byte{0x00}, []byte("Invalid_Tunn$l")),
			expectedErrorMsg: `invalid create_tunnel command: invalid name`,
		},
		"Listen_tunnel invalid payload": {
			indicator:        ListenTunnelIndicator,
			data:             []byte(""),
			expectedErrorMsg: `invalid payload: missing tunnel name`,
		},
		"Listen_tunnel invalid validation - Tunnel Name": {
			indicator:        ListenTunnelIndicator,
			data:             []byte("Mon_Tunn&l"),
			expectedErrorMsg: `invalid listen_tunnel command: invalid name`,
		},
		"Publish_message invalid payload": {
			indicator:        PublishMessageIndicator,
			data:             []byte(""),
			expectedErrorMsg: "invalid payload: missing separator, cannot determine values",
		},
		"Publish_message invalid validation - Tunnel Name": {
			indicator:        PublishMessageIndicator,
			data:             []byte("Inval+d_Tunnel Mon super message"),
			expectedErrorMsg: "invalid publish_message command: invalid tunnel_name",
		},
		"Publish_message invalid validation - Message": {
			indicator:        PublishMessageIndicator,
			data:             []byte("Bidule Invalide message chars &*&*"),
			expectedErrorMsg: "invalid publish_message command: invalid message",
		},
		"Receive_message invalid payload": {
			indicator:        ReceiveMessageIndicator,
			data:             []byte(""),
			expectedErrorMsg: "invalid payload: missing separator, cannot determine values",
		},
		"Receive_message invalid validation - Tunnel Name": {
			indicator:        ReceiveMessageIndicator,
			data:             []byte("Invalid&Tunnel Mon super message"),
			expectedErrorMsg: "invalid receive_message command: invalid tunnel_name",
		},
		"Receive_message invalid validation - Message": {
			indicator:        ReceiveMessageIndicator,
			data:             []byte("Bidule Invalide message chars &*&*"),
			expectedErrorMsg: "invalid receive_message command: invalid message",
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd, err := Parse(tc.indicator, "abcd1234", tc.data)
			assert.EqualError(t, err, tc.expectedErrorMsg)
			assert.Nil(t, cmd)
		})
	}

}
