package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAck_Validate(t *testing.T) {
	cmd := NewAck()
	assert.NoError(t, cmd.Validate())
}

func TestAck_Info(t *testing.T) {
	cmd := NewAck()
	assert.Equal(t, "ACK", cmd.Info())
}

func TestAck_ID(t *testing.T) {
	cmd := NewAck()
	assert.Equal(t, newIDValue, cmd.TransactionID())
}

func TestAck_Indicator(t *testing.T) {
	cmd := NewAck()
	assert.Equal(t, AcknowledgementIndicator, cmd.Indicator())
}

func TestAck_Data(t *testing.T) {
	cmd := NewAck()
	assert.Equal(t, []byte(ackData), cmd.Data())
}

func TestNack_Validate(t *testing.T) {
	cmd := NewNack()
	assert.NoError(t, cmd.Validate())
}

func TestNack_Info(t *testing.T) {
	cmd := NewNack()
	assert.Equal(t, "NACK", cmd.Info())
}

func TestNack_ID(t *testing.T) {
	cmd := NewNack()
	assert.Equal(t, newIDValue, cmd.TransactionID())
}

func TestNack_Indicator(t *testing.T) {
	cmd := NewNack()
	assert.Equal(t, AcknowledgementIndicator, cmd.Indicator())
}

func TestNack_Data(t *testing.T) {
	cmd := NewNack()
	assert.Equal(t, []byte(nackData), cmd.Data())
}
