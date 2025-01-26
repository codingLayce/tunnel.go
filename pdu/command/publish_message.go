package command

import (
	"bytes"
	"fmt"
	"regexp"
)

var messageValidator = regexp.MustCompile(`^[a-zA-Z0-9 _\d]+$`)

type PublishMessage struct {
	transactionID string

	TunnelName string
	Message    string
}

func parsePublishMessage(transactionID string, data []byte) (Command, error) {
	separatorIdx := bytes.Index(data, []byte(" "))
	if separatorIdx == -1 {
		return nil, fmt.Errorf("invalid payload: missing separator, cannot determine values")
	}

	tunnelName := data[:separatorIdx]
	message := data[separatorIdx+1:]

	cmd := NewPublishMessageWithTransactionID(transactionID, string(tunnelName), string(message))
	err := cmd.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid publish_message command: %s", err)
	}
	return cmd, nil
}

func NewPublishMessage(tunnelName, message string) *PublishMessage {
	return &PublishMessage{
		transactionID: newID(),
		TunnelName:    tunnelName,
		Message:       message,
	}
}

func NewPublishMessageWithTransactionID(transactionID, tunnelName, message string) *PublishMessage {
	return &PublishMessage{transactionID: transactionID,
		TunnelName: tunnelName,
		Message:    message,
	}
}

func (cmd *PublishMessage) Validate() error {
	if !tunnelNameValidator.MatchString(cmd.TunnelName) {
		return fmt.Errorf("invalid tunnel_name")
	}
	if !messageValidator.MatchString(cmd.Message) {
		return fmt.Errorf("invalid message")
	}
	return nil
}

func (cmd *PublishMessage) Info() string {
	return fmt.Sprintf("PUBLISH_MESSAGE[%s]message_size(%d)", cmd.TunnelName, len(cmd.Message))
}

func (cmd *PublishMessage) TransactionID() string {
	return cmd.transactionID
}

func (cmd *PublishMessage) Indicator() byte {
	return PublishMessageIndicator
}

func (cmd *PublishMessage) Data() []byte {
	buf := bytes.Buffer{}
	buf.WriteString(cmd.TunnelName)
	buf.WriteByte(' ')
	buf.WriteString(cmd.Message)
	return buf.Bytes()
}
