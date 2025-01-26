package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAck_Info(t *testing.T) {
	assert.Equal(t, "ACK", NewAck().Info())
}

func TestAck_TransactionID(t *testing.T) {
	assert.Equal(t, newIDValue, NewAck().TransactionID())
}

func TestAck_Indicator(t *testing.T) {
	assert.Equal(t, AcknowledgementIndicator, NewAck().Indicator())
}

func TestAck_Data(t *testing.T) {
	assert.Equal(t, []byte(ackData), NewAck().Data())
}

func TestNack_Info(t *testing.T) {
	assert.Equal(t, "NACK", NewNack().Info())
}

func TestNack_TransactionID(t *testing.T) {
	assert.Equal(t, newIDValue, NewNack().TransactionID())
}

func TestNack_Indicator(t *testing.T) {
	assert.Equal(t, AcknowledgementIndicator, NewNack().Indicator())
}

func TestNack_Data(t *testing.T) {
	assert.Equal(t, []byte(nackData), NewNack().Data())
}
