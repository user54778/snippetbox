package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
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

// Create a new helper, return a pointer to a templateData struct init w/ current year.
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear: time.Now().Year(),
		Flash:       app.sessionManager.PopString(r.Context(), "flash"),
		// auto-added every time we render a template
		IsAuthenticated: app.isAuthenticated(r),
	}
}

// A helper method to render templates from the in-memory cache.
// page is the template.
// data is template data we want to pass into the template to render.
func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	// Retrieve the relevant template set from the cache based on the page
	// name (i.e., 'home.tmpl'). If no entry exists in the cache, create a new
	// error and call serverError().
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	// Step 1: Initialize a buffer to trial render our template into.
	buf := new(bytes.Buffer)

	// Execute the template set and write the response body
	// Call serverError() upon any error.
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// If template writes to buffer with no errors, then we are safe to go ahead.
	w.WriteHeader(status)

	// Write contents of the buffer to our ResponseWriter.
	// NOTE: notice how this is another situation ResponseWriter takes an io.Writer.
	buf.WriteTo(w)
}

// dst is the target destination we want to decode the form data into.
func (app *application) decodePostForm(r *http.Request, dst any) error {
	// call PostForm() on the request
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		// Using an invalid target destination will result in an error, of type
		// *form.InvalidDecoderError, so use errors.As() to check for this.
		// This is a CRITICAL error if it occurs, so immediately panic the applicaiton.
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		// Return all other errors as normal
		return err
	}

	return nil
}

// Helper to check if a request is made by an authenticated user via
// checking existence of authenticatedUserID value in their session data.
func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "authenticatedUserID")
}
