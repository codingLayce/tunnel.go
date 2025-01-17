package command

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

func NewAck() *Ack                                      { return &Ack{transactionID: newID()} }
func NewAckWithTransactionID(transactionID string) *Ack { return &Ack{transactionID: transactionID} }
func (ack *Ack) Validate() error                        { return nil }
func (ack *Ack) Info() string                           { return "ACK" }
func (ack *Ack) TransactionID() string                  { return ack.transactionID }
func (ack *Ack) Indicator() byte                        { return AcknowledgementIndicator }
func (ack *Ack) Data() []byte                           { return []byte(ackData) }

func NewNack() *Nack                                      { return &Nack{transactionID: newID()} }
func NewNackWithTransactionID(transactionID string) *Nack { return &Nack{transactionID: transactionID} }
func (nack *Nack) Validate() error                        { return nil }
func (nack *Nack) Info() string                           { return "NACK" }
func (nack *Nack) TransactionID() string                  { return nack.transactionID }
func (nack *Nack) Indicator() byte                        { return AcknowledgementIndicator }
func (nack *Nack) Data() []byte                           { return []byte(nackData) }
