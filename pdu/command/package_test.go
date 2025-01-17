package command

import (
	"bytes"
	"testing"
)

const newIDValue = "abcd1234"

func TestMain(m *testing.M) {
	// Mocks the id generation function.
	newID = func() string { return newIDValue }

	m.Run()
}

// buildPayload creates the payload as follows: <indicator><data>
func buildPayload(indicator byte, data ...[]byte) []byte {
	buf := bytes.Buffer{}
	buf.WriteByte(indicator)
	for _, d := range data {
		buf.Write(d)
	}
	return buf.Bytes()
}
