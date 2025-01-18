package mock

import "testing"

// Do replace dest by value and reset it after test execution.
func Do[T any](t *testing.T, dest *T, value T) {
	cache := *dest
	*dest = value
	t.Cleanup(func() {
		*dest = cache
	})
}
