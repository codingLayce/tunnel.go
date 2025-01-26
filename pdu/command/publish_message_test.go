package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishMessage_Info(t *testing.T) {
	assert.Equal(t, "PUBLISH_MESSAGE[Bidule]message_size(4)", NewPublishMessage("Bidule", "toto").Info())
}

func TestPublishMessage_TransactionID(t *testing.T) {
	assert.Equal(t, newIDValue, NewPublishMessage("Bidule", "toto").TransactionID())
}

func TestPublishMessage_Indicator(t *testing.T) {
	assert.Equal(t, PublishMessageIndicator, NewPublishMessage("Bidule", "toto").Indicator())
}

func TestPublishMessage_Data(t *testing.T) {
	assert.Equal(t, data([]byte("Bidule"), []byte{' '}, []byte("toto")), NewPublishMessage("Bidule", "toto").Data())
}
