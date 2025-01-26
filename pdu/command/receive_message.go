package command

import (
	"bytes"
	"fmt"
)

type ReceiveMessage struct {
	transactionID string

	TunnelName string
	Message    string
}

func parseReceiveMessage(transactionID string, data []byte) (Command, error) {
	separatorIdx := bytes.Index(data, []byte(" "))
	if separatorIdx == -1 {
		return nil, fmt.Errorf("invalid payload: missing separator, cannot determine values")
	}

	tunnelName := data[:separatorIdx]
	message := data[separatorIdx+1:]

	cmd := NewReceiveMessageWithTransactionID(transactionID, string(tunnelName), string(message))
	err := cmd.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid receive_message command: %s", err)
	}
	return cmd, nil
}

func NewReceiveMessage(tunnelName, message string) *ReceiveMessage {
	return &ReceiveMessage{
		transactionID: newID(),
		TunnelName:    tunnelName,
		Message:       message,
	}
}

func NewReceiveMessageWithTransactionID(transactionID, tunnelName, message string) *ReceiveMessage {
	cmd := NewReceiveMessage(tunnelName, message)
	cmd.transactionID = transactionID
	return cmd
}

func (cmd *ReceiveMessage) Validate() error {
	if !tunnelNameValidator.MatchString(cmd.TunnelName) {
		return fmt.Errorf("invalid tunnel_name")
	}
	if !messageValidator.MatchString(cmd.Message) {
		return fmt.Errorf("invalid message")
	}
	return nil
}

func (cmd *ReceiveMessage) Info() string {
	return fmt.Sprintf("RECEIVE_MESSAGE[%s]message_size(%d)", cmd.TunnelName, len(cmd.Message))
}

func (cmd *ReceiveMessage) TransactionID() string {
	return cmd.transactionID
}

func (cmd *ReceiveMessage) Indicator() byte {
	return ReceiveMessageIndicator
}

func (cmd *ReceiveMessage) Data() []byte {
	buf := bytes.Buffer{}
	buf.WriteString(cmd.TunnelName)
	buf.WriteByte(' ')
	buf.WriteString(cmd.Message)
	return buf.Bytes()
}
