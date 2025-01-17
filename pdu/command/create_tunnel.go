package command

import (
	"bytes"
	"fmt"
	"regexp"
)

var createTunnelNameValidator = regexp.MustCompile(`^[a-zA-Z\d]+$`)

type TunnelType byte

const (
	BroadcastTunnel TunnelType = iota
)

type CreateTunnel struct {
	transactionID string

	Name string
	Type TunnelType
}

func parseCreateTunnel(transactionID string, data []byte) (Command, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("invalid payload: missing tunnel type")
	}

	cmd := NewCreateTunnel(string(data[1:]))
	cmd.Type = TunnelType(data[0])
	cmd.transactionID = transactionID
	err := cmd.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid create_tunnel command: %s", err)
	}
	return cmd, nil
}

func NewCreateTunnel(name string) *CreateTunnel {
	return &CreateTunnel{transactionID: newID(), Name: name}
}
func NewCreateTunnelWithTransactionID(transactionID, name string) *CreateTunnel {
	return &CreateTunnel{transactionID: transactionID, Name: name}
}
func (cmd *CreateTunnel) Validate() error {
	if cmd.Type != BroadcastTunnel {
		return fmt.Errorf("invalid type")
	}
	if createTunnelNameValidator.MatchString(cmd.Name) {
		return nil
	}
	return fmt.Errorf("invalid name")
}

func (cmd *CreateTunnel) Info() string          { return fmt.Sprintf("CREATE_TUNNEL(%s)", cmd.Name) }
func (cmd *CreateTunnel) TransactionID() string { return cmd.transactionID }
func (cmd *CreateTunnel) Indicator() byte       { return CreateTunnelIndicator }
func (cmd *CreateTunnel) Data() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte(byte(cmd.Type))
	buf.WriteString(cmd.Name)
	return buf.Bytes()
}
