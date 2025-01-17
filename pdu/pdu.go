package pdu

import (
	"bytes"
	"fmt"

	"github.com/codingLayce/tunnel.go/id"
	"github.com/codingLayce/tunnel.go/pdu/command"
)

const Delimiter byte = '\n'

func Marshal(cmd command.Command) []byte {
	buf := bytes.Buffer{}
	buf.WriteByte(cmd.Indicator())
	buf.WriteString(cmd.TransactionID())
	buf.Write(cmd.Data())
	buf.WriteByte(Delimiter)
	return buf.Bytes()
}

func Unmarshal(payload []byte) (command.Command, error) {
	if len(payload) < 10 { // 1 byte indicator + 8 bytes transactionID + 1 byte delimiter
		return nil, fmt.Errorf("invalid payload length: cannot be less than 10 bytes")
	}

	if payload[len(payload)-1] != Delimiter {
		return nil, fmt.Errorf("invalid payload delimiter %q, expected %q", payload[len(payload)-1], Delimiter)
	}

	indicator := payload[0]
	transactionID := string(payload[1:9])
	if !id.IsValid(transactionID) {
		return nil, fmt.Errorf("invalid transaction id")
	}
	data := payload[9 : len(payload)-1]

	return parseCommand(indicator, transactionID, data)
}

var parseCommand = command.Parse
