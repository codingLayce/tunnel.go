package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReceiveMessage_Info(t *testing.T) {
	assert.Equal(t, "RECEIVE_MESSAGE[Bidule]message_size(4)", NewReceiveMessage("Bidule", "toto").Info())
}

func TestReceiveMessage_TransactionID(t *testing.T) {
	assert.Equal(t, newIDValue, NewReceiveMessage("Bidule", "toto").TransactionID())
}

func TestReceiveMessage_Indicator(t *testing.T) {
	assert.Equal(t, ReceiveMessageIndicator, NewReceiveMessage("Bidule", "toto").Indicator())
}

func TestReceiveMessage_Data(t *testing.T) {
	assert.Equal(t, data([]byte("Bidule"), []byte{' '}, []byte("toto")), NewReceiveMessage("Bidule", "toto").Data())
}
