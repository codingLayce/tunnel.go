package id

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	rex, err := regexp.Compile(`[a-zA-Z0-9]{8}`)
	require.NoError(t, err)

	for i := 0; i < 100; i++ {
		assert.True(t, rex.MatchString(New()))
	}
}

func TestIsValid(t *testing.T) {
	for name, tc := range map[string]struct {
		id      string
		isValid bool
	}{
		"Empty": {
			id:      "",
			isValid: false,
		},
		"Length 1": {
			id:      "1",
			isValid: false,
		},
		"Length 2": {
			id:      "1a",
			isValid: false,
		},
		"Length 3": {
			id:      "1a2",
			isValid: false,
		},
		"Length 4": {
			id:      "1a2B",
			isValid: false,
		},
		"Length 5": {
			id:      "1a2b3",
			isValid: false,
		},
		"Length 6": {
			id:      "1a2b3C",
			isValid: false,
		},
		"Length 7": {
			id:      "1A2b3c4",
			isValid: false,
		},
		"Length 8": {
			id:      "1A2b3c4D",
			isValid: true,
		},
		"Length 9": {
			id:      "1A2b3c4D5",
			isValid: false,
		},
		"With space": {
			id:      "1A2b c4D",
			isValid: false,
		},
		"With special character": {
			id:      "1A2b_c4D",
			isValid: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.isValid, IsValid(tc.id))
		})
	}
}

// goos: linux
// goarch: amd64
// pkg: tunnel/id
// cpu: AMD Ryzen 7 7700 8-Core Processor
// BenchmarkNew
// BenchmarkNew-16    	 1749612	       657.7 ns/op
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}
