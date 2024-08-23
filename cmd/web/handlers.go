package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"snippetbox.adpollak.net/internal/models"
)

// Handler
// NOTE: This signature was changed to be defined as a method against the *application type.
// This allows us to not depend on some specific type.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Let's enforce a restriction on our servemux by 404ing if we ARE NOT on / exactly.
	if r.URL.Path != "/" {
		// http.NotFound(w, r)
		app.notFound(w)
		return
	}

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	for _, snippet := range snippets {
		fmt.Fprintf(w, "%+v\n", snippet)
	}
	/*
		// Let's now add functionality to render a template file containing an HTML page.
		// We will need to parse the template file.
		// Use template.ParseFiles() to read the template file into a template set (ts).
		//
		// The file path passed in MUST either
		// a) be relative to current working directory, or
		// b) be an absolute path
		//ts, err := template.ParseFiles("./ui/html/pages/home.tmpl")
		files := []string{
			"./ui/html/base.tmpl",
			"./ui/html/pages/home.tmpl",
			"./ui/html/partials/nav.tmpl",
		}

			ts, err := template.ParseFiles(files...)
			if err != nil {
				// log.Println(err.Error())
				// Write the log message to the applicaiton type instead to use the error logger.
				// app.errorLog.Println(err.Error())
				// http.Error(w, "Internal Server Error", 500)
				app.serverError(w, err)
				return
			}

			// We then use Execute() on the ts to write the template content as the response body.
			// The last parameter to Execute() represents any dynamic data we want to pass in.
			//
			// (Now ExecuteTemplate()) to write the content of the "base" template as the response body
			err = ts.ExecuteTemplate(w, "base", nil)
			if err != nil {
				// Write the log message to the applicaiton type instead to use the error logger.
				// app.errorLog.Println(err.Error())
				// log.Println(err.Error())
				// http.Error(w, "Internal Server Error", 500)
				app.serverError(w, err)
			}
	*/
}

// Handler
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// Let's update this handler to acept an id query string parameter from the user.
	// For now, (since no db), we will need to read the value of the id parameter and interpolate it
	// with a placeholder response.
	// Two steps to this:
	// 1) We need to retrieve the value of the id parameter from the URL query string
	// 2) Since this id IS untrusted user input, we should validate it to ensure its sensible.
	//    For our case, this just means it contains a positive integer.
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		//.http.NotFound(w, r)
		app.notFound(w)
		return
	}

	// w.Write([]byte("Display a specific snippet...\n"))

	// Use SnippetModel's Get method to retrieve the data for a specific
	// record based on its ID. If no record is found, return a 404 Not Found response.
	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// Interpolate the id parameter with a placeholder response.
	// fmt.Fprintf(w, "Display a specific snippet with ID %d...\n", id)

	// Write the snippet data as plain-text HTTP response body.
	// TODO: look at format specifier
	fmt.Fprintf(w, "%+v", snippet)
}

// Handler
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Now, let's restrict this route to only respond to HTTP requests using the POST method.
	// NOTE: how? -> Let's send a 405 (method not allowed) status code UNLESS the request is a POST.

	if r.Method != http.MethodPost {
		// Send a 405 status code -> How? Well, how is the ResponseWriter interface setup?
		// type ResponseWriter interface {
		//      Header() http.Header
		//      Write([]byte) (int, error)
		//      WriteHeader(statusCode int)
		// }
		// These methods MUST be called in a specific order:
		// 1) Header() is called to set response headers (if needed, otherwise don't call it)
		// 2) WriteHeader(statusCode int) with the HTTP status code for the response (unless sending a response with 200 status code)
		// 3) Write([]byte) is called to set the body for the response.
		//
		// Let's include an Allow header with our 405 response to let the user know what IS allowed.
		w.Header().Set("Allow", "POST")

		/* The http.Error() does the same as this in a more concise manner (send non-200 status code and plain-text response body)
		w.WriteHeader(405)
		w.Write([]byte("Method not allowed\n"))
		*/
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	// Create some variables holding dummy data.
	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	expires := 7

	// Pass the data to the SnippetModel.Insert() method, receiving the ID
	// of the new record back.
	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// w.Write([]byte("Create a new snippet...\n"))
	// Redirect the user to the relevant page for the snippet.
	// TODO: look up what StatusSeeOther is.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view?id=%d", id), http.StatusSeeOther)
}
