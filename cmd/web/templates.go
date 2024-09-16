package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"snippetbox.adpollak.net/internal/models"
	"snippetbox.adpollak.net/ui"
)

// Define a type to act as a holding structure for
// any dynamic data we want to pass to our HTML templates.
type templateData struct {
	CurrentYear     int // common dyn data we want to include on every page
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	Form            any // used to pass validation errors and prev submitted data back to template when re-display the form
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

// A function to cache our parsed tmpl files.
// We use an in-memory map with the type map[string]*template.Template to
// cache the parsed templates.
func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize new map
	cache := map[string]*template.Template{}

	// Use filepath.Glob() to get a slice of all filepaths that match the pattern
	// "./ui/html/pages/*.tmpl". This will give us a slice of all the filepaths
	// for our application 'page' templates such as [ui/html/pages/home.tmpl ui/html/pages/view.tmpl]
	// pages, err := filepath.Glob("./ui/html/pages/*.tmpl")
	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Loop through the page filepaths one-by-one
	for _, page := range pages {
		// Extract the filename (i.e., 'home.tmpl', etc) from full filepath
		// and assign it to the name variable
		name := filepath.Base(page)

		// Create a slice containing filepath patterns for the
		// templates we want to parse from fs.Glob().
		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		// Changed to parse the base template file into a template set
		// Changed again to register the FuncMap with the template set. This must be done
		// prior to calling ParseFiles. To do so, we use template.New() to create a empty
		// template set and then use Funcs() to register the template.FuncMap() and then parse
		// the file as normal.
		// ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl")

		// NOTE: Use ParseFS instead of ParseFiles to parse the template files
		// from the ui.Files embedded filesystem.
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		/*
			// Call ParseGlob() on this template set to add the page template.
			ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl")
			if err != nil {
				return nil, err
			}
		*/

		/*
			// Call ParseFiles() on this template set to add the page template.
			ts, err = ts.ParseFiles(page)
			if err != nil {
				return nil, err
			}
		*/

		// Add the template set to the map, using the name of the page,
		// (such as 'home.tmpl') as the key
		cache[name] = ts
	}

	return cache, nil
}

// NOTE: custom function used in our Go template.
// Create a humanDate function which returns a formateed string representation
// of a time.Time object.
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// Initialize a template.FuncMap object and store it in a global var.
// This is essentially a string-keyed map which acts as a lookup between the names
// of our custom template functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
}
