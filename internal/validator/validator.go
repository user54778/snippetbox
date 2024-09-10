package validator

import (
	"strings"
	"unicode/utf8"
)

// Contains a map of validation errors of our form fields.
type Validator struct {
	FieldErrors map[string]string
}

// NOTE: the three methods are for generally adding errors to the FieldErrors map.

// Return true if FieldErrors empty
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
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
func PermittedInt(value int, permittedValues ...int) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}
