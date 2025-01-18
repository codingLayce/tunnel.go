package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDo(t *testing.T) {
	a := 17
	assert.Equal(t, 17, a) // Default value
	t.Run("Other", func(t *testing.T) {
		Do(t, &a, 15)
		assert.Equal(t, 15, a) // Mock value
	})
	assert.Equal(t, 17, a) // Value is reset after cleanup
}
