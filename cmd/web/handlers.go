package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"snippetbox.adpollak.net/internal/models"

	"github.com/julienschmidt/httprouter"
)

// Represent the form data and validation errors for the
// form fields. Note all struct fields are deliberately EXPORTED
// (i.e., start w/ Capital letter). Struct fields must be exported
// in order to be read by the html/template package when rendering a template.
type snippetCreateForm struct {
	Title       string
	Content     string
	Expires     int
	FieldErrors map[string]string
}

// Handler
// NOTE: This signature was changed to be defined as a method against the *application type.
// This allows us to not depend on some specific type.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Before, we had to check the r.URL.Path != "/" for this handler.
	// Since httprouter matches the "/" path EXACTLY, we can remove this.

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Call newTemplateData() helper to get a templateData struct containing the
	// 'default' data (for now curr year) and add the snippets slice to it.
	data := app.newTemplateData(r)
	data.Snippets = snippets

	// Pass data to the render() helper
	app.render(w, http.StatusOK, "home.tmpl", data)
}

// Handler
func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// When parsing a request, any named parameters will be stored in the request context.
	params := httprouter.ParamsFromContext(r.Context())

	// Instead, use ByName() to get the value of id named parameter from the slice
	// and validate it as normal.
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.notFound(w)
		return
	}

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

	data := app.newTemplateData(r)
	data.Snippet = snippet

	// Use our new render helper.
	app.render(w, http.StatusOK, "view.tmpl", data)
}

// For now returns a placeholder response.
// Will show the html form.
func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("Display the form for creating a new snippet here..."))
	data := app.newTemplateData(r)

	app.render(w, http.StatusOK, "create.tmpl", data)
}

// Processes and uses form data of snippet. Upon completion
// redirects user to the view page to view their snippet.
func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// Step 1: parse request body
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Step 2: retrieve PostForm() data. Expires is a integer
	// so convert with strconv.Atoi() and throw an error is fail
	// conversion.
	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := snippetCreateForm{
		Title:       r.PostForm.Get("title"),
		Content:     r.PostForm.Get("content"),
		Expires:     expires,
		FieldErrors: map[string]string{},
	}

	// 1) title value not empty and <= 100 chars long
	if strings.TrimSpace(form.Title) == "" {
		form.FieldErrors["title"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(form.Title) > 100 {
		form.FieldErrors["title"] = "This field cannot be more than 100 characters"
	}

	// 1a) content value non-empty
	if strings.TrimSpace(form.Content) == "" {
		form.FieldErrors["content"] = "This field cannot be blank"
	}

	// 2) expires value matches one of our permitted values
	// 1, 7, 365
	if expires != 1 && expires != 7 && expires != 365 {
		form.FieldErrors["expires"] = "This field must equal 1, 7, or 365"
	}

	// If any validation errors occur:
	// a) Re-display web-page form (create.tmpl) passing in snippetCreateForm instance
	// as dyn data in the Form field.
	// b) re-populate any prev submitted data, done via passing in snippetCreateForm instance to Form field.
	if len(form.FieldErrors) > 0 {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	// insert title, content, expiration into db
	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// w.Write([]byte("Create a new snippet...\"))
	// Redirect the user to the relevant page for the snippet.
	// Update redirect path to the new clean URL format.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
