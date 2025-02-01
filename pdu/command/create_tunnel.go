package command

import (
	"bytes"
	"fmt"
	"regexp"
)

var tunnelNameValidator = regexp.MustCompile(`^[a-zA-Z_.\-\d]+$`)

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

	cmd := NewCreateTunnelWithTransactionID(transactionID, string(data[1:]))
	cmd.Type = TunnelType(data[0])
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
	cmd := NewCreateTunnel(name)
	cmd.transactionID = transactionID
	return cmd
}

func (cmd *CreateTunnel) Validate() error {
	if cmd.Type != BroadcastTunnel {
		return fmt.Errorf("invalid type")
	}
	if !tunnelNameValidator.MatchString(cmd.Name) {
		return fmt.Errorf("invalid name")
	}
	// TODO : Limit name size
	return nil
}

func (cmd *CreateTunnel) Info() string {
	// TODO : Specify the Tunnel type
	return fmt.Sprintf("CREATE_TUNNEL(%s)", cmd.Name)
}
func (cmd *CreateTunnel) TransactionID() string { return cmd.transactionID }
func (cmd *CreateTunnel) Indicator() byte       { return CreateTunnelIndicator }
func (cmd *CreateTunnel) Data() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte(byte(cmd.Type))
	buf.WriteString(cmd.Name)
	return buf.Bytes()
}
