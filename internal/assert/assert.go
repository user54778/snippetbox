package assert

import "testing"

func Equal[T comparable](t *testing.T, actual, expected T) {
	// NOTE: Helper() here indicates that our Equal() function is a test helper.
	// When t.Errorf() is called from Equal(), Go will report the filename
	// and line number of the code that called our Equal() function in the output.
	t.Helper()

	if actual != expected {
		t.Errorf("actual: %v; expected: %v", actual, expected)
	}
}
