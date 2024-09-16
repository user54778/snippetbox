package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Contains a map of validation errors of our form fields.
// Hold any validation errors not related to a specific form field.
type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

// NOTE: the three methods are for generally adding errors to the FieldErrors map.

// Return true if FieldErrors empty
func (v *Validator) Valid() bool {
	// Updated to check NonFieldErrors slice also empty
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// Add an error message to the FieldErrors map (as long as no entry already
// exists for the given key)
func (v *Validator) AddFieldError(key, message string) {
	// Init map first if not init'd
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	// Get the value from the map, if it exists add the error
	// message to the FieldErrors map
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// Helper to add error messages to the NonFieldErrors slice.
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// Add an error message to the FieldErrors map only if validation
// check is not ok.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NOTE: These functions help us perform SPECIFIC validation checks
// (hence not methods) (i.e., our original validation checks)

// Check non-empty string
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// Check if a value contains no more than n characters
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// Check if a value is in a list of permitted integers.
// Updated to instead check for a list of T values.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}

// Use regexp.MustCompile() to parse a regex pattern for sanity checking the format of an email address.
// Returns a pointer to a compiled regexp.Regexp type, or panics in the event of an error.
// Parsing this pattern once at startup and storing the compiled *regexp.Regexp
// in a variable is more performant than re-parsing the pattern each time we need it.
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Returns true if a value contains at least n characters.
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// Returns true if a value matches the provided compiled
// regex pattern.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
