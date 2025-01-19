package command

import (
	"bytes"
	"fmt"
)

type ListenTunnel struct {
	transactionID string

	Name string
}

func NewListenTunnel(name string) *ListenTunnel {
	return &ListenTunnel{
		transactionID: newID(),
		Name:          name,
	}
}

func NewListenTunnelWithTransactionID(transactionID, name string) *ListenTunnel {
	return &ListenTunnel{
		transactionID: transactionID,
		Name:          name,
	}
}

func parseListenTunnel(transactionID string, data []byte) (Command, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("invalid payload: missing tunnel name")
	}

	cmd := NewListenTunnel(string(data))
	cmd.transactionID = transactionID
	err := cmd.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid listen_tunnel command: %s", err)
	}
	return cmd, nil
}

func (cmd *ListenTunnel) Validate() error {
	if createTunnelNameValidator.MatchString(cmd.Name) {
		return nil
	}
	return fmt.Errorf("invalid name")
}

func (cmd *ListenTunnel) Info() string {
	// TODO : Specify the Tunnel type
	return fmt.Sprintf("LISTEN_TUNNEL(%s)", cmd.Name)
}
func (cmd *ListenTunnel) TransactionID() string { return cmd.transactionID }
func (cmd *ListenTunnel) Indicator() byte       { return ListenTunnelIndicator }
func (cmd *ListenTunnel) Data() []byte {
	buf := bytes.Buffer{}
	buf.WriteString(cmd.Name)
	return buf.Bytes()
}
