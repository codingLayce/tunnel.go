package command

import (
	"fmt"

	"github.com/codingLayce/tunnel.go/id"
)

const (
	AcknowledgementIndicator byte = '@'
	CreateTunnelIndicator    byte = '+'
	ListenTunnelIndicator    byte = '#'
	PublishMessageIndicator  byte = '>'
)

type Command interface {
	Validate() error
	Info() string
	TransactionID() string
	Indicator() byte
	Data() []byte
}

// Parse to the corresponding command.
// Returns an error if no command fit the given attributes or if it's invalid.
func Parse(indicator byte, transactionID string, data []byte) (Command, error) {
	switch indicator {
	case AcknowledgementIndicator:
		return parseAcknowledgement(transactionID, data)
	case CreateTunnelIndicator:
		return parseCreateTunnel(transactionID, data)
	case ListenTunnelIndicator:
		return parseListenTunnel(transactionID, data)
	case PublishMessageIndicator:
		return parsePublishMessage(transactionID, data)
	default:
		return nil, fmt.Errorf("invalid command indicator: unknown 0x%x", indicator)
	}
}

var newID = id.New
