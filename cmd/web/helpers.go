package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// Write an error message and stack trace to the errorLog, then send
// a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	// Gets the stack trace for the current goroutine and appends it to the log msg.
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace) // Use our logger's output function

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Sends a specific status code and corresponding description to the user.
// Used later in the book to send responses such as 400 "Bad Request" upon issue
// with the request sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	// NOTE: statusText generates a human-friendly text representation of a given HTTP status code.
	http.Error(w, http.StatusText(status), status)
}

// A convenience wrapper around clientError which sends a 404 Not Found response
// to the user.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
