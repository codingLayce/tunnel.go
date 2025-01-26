package command

import "fmt"

const (
	ackData  = "OK"
	nackData = "KO"
)

type (
	// Ack represents a valid acknowledgement.
	Ack struct {
		transactionID string
	}

	// Nack represents a non acknowledgement.
	Nack struct {
		transactionID string
	}
)

func parseAcknowledgement(transactionID string, data []byte) (Command, error) {
	var cmd Command
	switch string(data) {
	case ackData:
		cmd = NewAckWithTransactionID(transactionID)
	case nackData:
		cmd = NewNackWithTransactionID(transactionID)
	default:
		return nil, fmt.Errorf("invalid acknowledgement command: unknown data %q", string(data))
	}

	_ = cmd.Validate() // Currently validation cannot fail for ack / nack. Keeping it because will be useful and make tests coverage.
	return cmd, nil
}

func NewAck() *Ack { return &Ack{transactionID: newID()} }
func NewAckWithTransactionID(transactionID string) *Ack {
	cmd := NewAck()
	cmd.transactionID = transactionID
	return cmd
}
func (ack *Ack) Validate() error       { return nil }
func (ack *Ack) Info() string          { return "ACK" }
func (ack *Ack) TransactionID() string { return ack.transactionID }
func (ack *Ack) Indicator() byte       { return AcknowledgementIndicator }
func (ack *Ack) Data() []byte          { return []byte(ackData) }

func NewNack() *Nack { return &Nack{transactionID: newID()} }
func NewNackWithTransactionID(transactionID string) *Nack {
	cmd := NewNack()
	cmd.transactionID = transactionID
	return cmd
}
func (nack *Nack) Validate() error       { return nil }
func (nack *Nack) Info() string          { return "NACK" }
func (nack *Nack) TransactionID() string { return nack.transactionID }
func (nack *Nack) Indicator() byte       { return AcknowledgementIndicator }
func (nack *Nack) Data() []byte          { return []byte(nackData) }
