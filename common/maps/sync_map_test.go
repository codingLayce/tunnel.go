package maps

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncMap(t *testing.T) {
	m := NewSyncMap[int, string]()

	concurrentProcesses := 10_000

	wg := sync.WaitGroup{}
	wg.Add(concurrentProcesses)
	for i := 0; i < concurrentProcesses; i++ {
		go func() {
			defer wg.Done()
			m.Put(i, fmt.Sprintf("Value %d", i))
		}()
	}
	wg.Wait()

	assert.Equal(t, concurrentProcesses, m.Len())

	wg.Add(concurrentProcesses)
	for i := 0; i < concurrentProcesses; i++ {
		go func() {
			defer wg.Done()
			assert.True(t, m.Has(i))
			value, ok := m.Get(i)
			assert.True(t, ok)
			assert.Equal(t, fmt.Sprintf("Value %d", i), value)
			m.Delete(i)
		}()
	}
	wg.Wait()

	assert.Equal(t, 0, m.Len())
}

func TestSyncMap_Iterator(t *testing.T) {
	m := NewSyncMap[int, string]()

	for i := 0; i < 100; i++ {
		m.Put(i, fmt.Sprintf("Value %d", i))
	}

	concurrentProcesses := 100
	wg := sync.WaitGroup{}
	wg.Add(concurrentProcesses)
	for i := 0; i < concurrentProcesses; i++ {
		go func() {
			defer wg.Done()
			for key, value := range m.Iterator() {
				assert.Equal(t, fmt.Sprintf("Value %d", key), value)
			}
		}()
	}
	wg.Wait()
}
