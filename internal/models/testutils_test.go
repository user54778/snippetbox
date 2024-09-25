package models

import (
	"database/sql"
	"os"
	"testing"
)

// Database used for model integration tests.
func newTestDB(t *testing.T) *sql.DB {
	// Creates a new *sql.DB connection pool
	// Use multiStatements=true since our setup and teardown scripts
	// contain multiple SQL statements.
	db, err := sql.Open("mysql", "test_web:pass@/test_snippetbox?parseTime=true&multiStatements=true")
	if err != nil {
		t.Fatal(err)
	}

	// Read setup SQL script from file and execute the statements
	// Executes the setup.SQL script
	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	// Registers a cleanup function which executes
	// the teardown.sql script and closes the connection pool
	t.Cleanup(func() {
		script, err := os.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}

		db.Close()
	})

	return db
}
