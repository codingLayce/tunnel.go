package command

import (
	"fmt"
	"regexp"
)

var createTunnelNameValidator = regexp.MustCompile(`^[a-zA-Z\d]+$`)

type CreateTunnel struct {
	transactionID string

	Name string
}

func NewCreateTunnel(name string) *CreateTunnel {
	return &CreateTunnel{transactionID: newID(), Name: name}
}
func NewCreateTunnelWithTransactionID(transactionID, name string) *CreateTunnel {
	return &CreateTunnel{transactionID: transactionID, Name: name}
}
func (cmd *CreateTunnel) Validate() error {
	if createTunnelNameValidator.MatchString(cmd.Name) {
		return nil
	}
	return fmt.Errorf("invalid name")
}

func (cmd *CreateTunnel) Info() string          { return fmt.Sprintf("CREATE_TUNNEL(%s)", cmd.Name) }
func (cmd *CreateTunnel) TransactionID() string { return cmd.transactionID }
func (cmd *CreateTunnel) Indicator() byte       { return CreateTunnelIndicator }
func (cmd *CreateTunnel) Data() []byte          { return []byte(cmd.Name) }
