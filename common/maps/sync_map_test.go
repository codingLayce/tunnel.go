package maps

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncMap(t *testing.T) {
	m := NewSyncMap[int, string]()

	parallelism := 10_000

	wg := sync.WaitGroup{}
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func() {
			defer wg.Done()
			m.Put(i, fmt.Sprintf("Value %d", i))
		}()
	}
	wg.Wait()

	assert.Equal(t, parallelism, m.Len())

	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func() {
			defer wg.Done()
			value, ok := m.Get(i)
			assert.True(t, ok)
			assert.Equal(t, fmt.Sprintf("Value %d", i), value)
			m.Delete(i)
		}()
	}
	wg.Wait()

	assert.Equal(t, 0, m.Len())
}
