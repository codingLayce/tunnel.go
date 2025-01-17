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

func data(args ...[]byte) []byte {
	buf := bytes.Buffer{}
	for _, arg := range args {
		buf.Write(arg)
	}
	return buf.Bytes()
}
