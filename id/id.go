package id

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

var (
	generator = rand.New(rand.NewSource(time.Now().Unix()))
	mtx       = sync.Mutex{}
)

const (
	validCharacters = "abcdefghijklmnopqrstuvwxyz0123456789"
	nbPossibleChars = 36
	idLength        = 8
)

// New generates an id made of 8 bytes.
// There is 2.821.109.907.456 possible ids.
func New() string {
	mtx.Lock()
	defer mtx.Unlock()
	builder := strings.Builder{}
	for i := 0; i < idLength; i++ {
		builder.WriteByte(validCharacters[generator.Intn(nbPossibleChars)])
	}
	return builder.String()
}

// IsValid determines if the given id is a valid one.
func IsValid(id string) bool {
	if len(id) != idLength {
		return false
	}
	for _, character := range id {
		if !strings.Contains(validCharacters, string(character)) {
			return false
		}
	}
	return true
}
