package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func home(w http.ResponseWriter, r *http.Request) {
	// Let's enforce a restriction on our servemux by 404ing if we ARE NOT on / exactly.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

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
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// We then use Execute() on the ts to write the template content as the response body.
	// The last parameter to Execute() represents any dynamic data we want to pass in.
	//
	// (Now ExecuteTemplate()) to write the content of the "base" template as the response body
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

// Handler
func snippetView(w http.ResponseWriter, r *http.Request) {
	// Let's update this handler to acept an id query string parameter from the user.
	// For now, (since no db), we will need to read the value of the id parameter and interpolate it
	// with a placeholder response.
	// Two steps to this:
	// 1) We need to retrieve the value of the id parameter from the URL query string
	// 2) Since this id IS untrusted user input, we should validate it to ensure its sensible.
	//    For our case, this just means it contains a positive integer.
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	// w.Write([]byte("Display a specific snippet...\n"))
	// Interpolate the id parameter with a placeholder response.
	fmt.Fprintf(w, "Display a specific snippet with ID %d...\n", id)
}

// Handler
func snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Now, let's restrict this route to only respond to HTTP requests using the POST method.
	// NOTE: how? -> Let's send a 405 (method not allowed) status code UNLESS the request is a POST.

	if r.Method != "POST" {
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
		http.Error(w, "Method not allowed\n", http.StatusMethodNotAllowed) // == 405
		return
	}

	w.Write([]byte("Create a new snippet...\n"))
}
