package assert

import (
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	// NOTE: Helper() here indicates that our Equal() function is a test helper.
	// When t.Errorf() is called from Equal(), Go will report the filename
	// and line number of the code that called our Equal() function in the output.
	t.Helper()

	if actual != expected {
		t.Errorf("actual: %v; expected: %v", actual, expected)
	}
}

// Used to check if response body of HTTP response contains some specific
// content.
func StringContains(t *testing.T, actual, expectedSubstring string) {
	t.Helper()

	if !strings.Contains(actual, expectedSubstring) {
		t.Errorf("actual: %q; expected: %q", actual, expectedSubstring)
	}
}

// Assertion to check for nil error value.
func NilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("actual: %v; expected: nil", actual)
	}
}
