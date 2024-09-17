package main

import (
	"testing"
	"time"

	"snippetbox.adpollak.net/internal/assert"
)

func TestHumanDate(t *testing.T) {
	/*
		tm := time.Date(2022, 3, 17, 10, 15, 0, 0, time.UTC)
		hd := humanDate(tm)

		// Check output from humanDate in correct format.
		if hd != "17 Mar 2022 at 10:15" {
			t.Errorf("got %q; want %q", hd, "17 Mar 2022 at 10:15")
		}
	*/

	// Create the table
	tests := []struct {
		name     string
		tm       time.Time
		expected string
	}{
		{
			name:     "UTC",
			tm:       time.Date(2022, 3, 17, 10, 15, 0, 0, time.UTC),
			expected: "17 Mar 2022 at 10:15",
		},
		{
			name:     "Empty",
			tm:       time.Time{},
			expected: "",
		},
		{
			name:     "CET",
			tm:       time.Date(2022, 3, 17, 10, 15, 0, 0, time.FixedZone("CET", 1*60*60)),
			expected: "17 Mar 2022 at 09:15",
		},
	}
	// Loop over the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)
			/*
				if hd != tt.expected {
					t.Errorf("actual %q; expected %q", hd, tt.expected)
				}
			*/
			// Instead, use our new Equal() helper from our assert pkg
			// to compare the expected and actual values.
			assert.Equal(t, hd, tt.expected)
		})
	}
}
