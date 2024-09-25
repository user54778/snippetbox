package models

import (
	"testing"

	"snippetbox.adpollak.net/internal/assert"
)

// An integration test for the Exists() method.
func TestUserModelExists(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration tests")
	}

	tests := []struct {
		name   string
		userID int
		want   bool
	}{
		{
			name:   "Valid ID",
			userID: 1,
			want:   true,
		},
		{
			name:   "Zero ID",
			userID: 0,
			want:   false,
		},
		{
			name:   "Non-existent ID",
			userID: 2,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call newTestDB() helper to get a connection pool to our test db.
			// Calling inside Run means it will be setup and torn down per sub-test.
			db := newTestDB(t)

			// New UserModel instance
			m := UserModel{db}

			// Call UserModel.Exists() to check return value and err match
			// the expected values for the sub-tests.
			exists, err := m.Exists(tt.userID)

			assert.Equal(t, exists, tt.want)
			assert.NilError(t, err)
		})
	}
}
