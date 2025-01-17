package pdu

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/codingLayce/tunnel.go/pdu/command"
)

type FakeCommand struct {
	id        string
	indicator byte
	data      []byte
}

func (f FakeCommand) Validate() error       { return nil }
func (f FakeCommand) Info() string          { return "" }
func (f FakeCommand) TransactionID() string { return f.id }
func (f FakeCommand) Indicator() byte       { return f.indicator }
func (f FakeCommand) Data() []byte          { return f.data }

func TestMarshal(t *testing.T) {
	payload := Marshal(FakeCommand{
		id:        "abcd1234",
		indicator: '=',
		data:      []byte("data"),
	})
	assert.Equal(t, []byte("=abcd1234data\n"), payload)
}

func TestUnmarshal(t *testing.T) {
	// Mocks parse command
	parseCommand = func(indicator byte, transactionID string, data []byte) (command.Command, error) {
		assert.Equal(t, byte('='), indicator)
		assert.Equal(t, "toto1234", transactionID)
		assert.Equal(t, []byte("MyData"), data)
		return nil, nil
	}
	t.Cleanup(func() { parseCommand = command.Parse })

	cmd, err := Unmarshal([]byte("=toto1234MyData\n"))
	require.NoError(t, err)
	require.Nil(t, cmd)
}

func TestUnmarshal_Error(t *testing.T) {
	for name, tc := range map[string]struct {
		payload              []byte
		expectedErrorMessage string
	}{
		"Nil payload": {
			payload:              nil,
			expectedErrorMessage: "invalid payload length: cannot be less than 10 bytes",
		},
		"Empty payload": {
			payload:              []byte{},
			expectedErrorMessage: "invalid payload length: cannot be less than 10 bytes",
		},
		"Payload without transaction id": {
			payload:              []byte("=MyData\n"),
			expectedErrorMessage: "invalid payload length: cannot be less than 10 bytes",
		},
		"Payload with invalid transaction id": {
			payload:              []byte("=abcd123_\n"),
			expectedErrorMessage: "invalid transaction id",
		},
		"Payload with invalid delimiter": {
			payload:              []byte("=abcd1234\r"),
			expectedErrorMessage: `invalid payload delimiter '\r', expected '\n'`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			cmd, err := Unmarshal(tc.payload)
			assert.EqualError(t, err, tc.expectedErrorMessage)
			assert.Nil(t, cmd)
		})
	}
}
