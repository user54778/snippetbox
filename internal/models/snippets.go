package models

import (
	"database/sql"
	"errors"
	"time"
)

// Hold the data for an individual snippet.
// Notice that the fields corresponds to fields in the MySQL
// table.
type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// Define a SnippetModel type which wraps a sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

// Insert a new snippet into the database.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	// The SQL statement we want to execute.
	// We use ? to indicate placeholder parameters for data we want to insert into the database.
	// As the data is untrusted user input, we'd rather do this than interpolate data in the query.
	// NOTE: `` is used since we split the string into multiple lines.
	stmt := `INSERT INTO snippets (title, content, created, expires)
  VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// Use the Exec() method on the embedded connection pool to execute the statement.
	// Takes in a SQL statement, followed by additional info for the query.
	// Returns a sql.Result type, which contains basic information about what happened when the
	// statement was executed.
	// NOTE: it is common to ignore the sql.Result return value if not needed.
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// Use the LastInsertId() method on the result to get the ID of our
	// newly inserted record in the snippets table.
	// NOTE: not all drivers and dbs support this method; for example, postgres does NOT.
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Return a specific (single) snippet based on its id.
func (m *SnippetModel) Get(id int) (*Snippet, error) {
	// The SQL statement we want to execute.
	stmt := `SELECT id, title, content, created, expires FROM snippets
  WHERE expires > UTC_TIMESTAMP() AND id = ?`

	// Use QueryRow() on the connection pool to execute our SQL statement, passing in the
	// untrusted* id variable as the value for the placeholder parameter.
	// This returns a pointer to a sql.Row object which holds the result from the database.
	tuple := m.DB.QueryRow(stmt, id)

	// Initialize a pointer to a new zeroed Snippet
	s := &Snippet{}

	// Copy the values from each field in sql.Row to the corresponding field in the Snippet.
	// Notice that arguments are pointers to the place you want to copy data to; we want to copy the
	// pointer to the location of the data, NOT copy the value.
	err := tuple.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// Scenario: The query returns no tuples, in which case row.Scan()
		// will return a sql.ErrNoRows error.
		// Use errors.Is() to check for that error specifically, and return a custom
		// ErrNoRecord instead.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	// Everything OK, return Snippet
	return s, nil
}

// Return the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]*Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
  WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	// Returns a resultset containg result of our query.
	tuples, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	// Ensure resultset is properly closed; this should come after the error check on Query
	// otherwise, if it returns an error it will panic trying to close a nil resultset.
	defer tuples.Close()

	snippets := []*Snippet{}

	// Iterate over the tuples in the resultset. Prepares first and each subsequent
	// tuple to be acted on by the Scan() method. If iteration over all rows completes,
	// the resultset automatically closes itself and frees-up the underlying database connection.
	for tuples.Next() {
		s := &Snippet{}

		err := tuples.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	// Retrieve any potential error that occurred during iteration.
	if err = tuples.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
