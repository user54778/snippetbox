package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

// Returns a ServeMux containing our application routes.
// NOTE: we changed the return type from ServeMux to Handler
// since we wanted to wrap our middleware around the ServeMux.
func (app *application) routes() http.Handler {
	// Initialize the router
	router := httprouter.New()
	// mux := http.NewServeMux()

	// Create a handler function that wraps our notFound helper, and then assigns it as the custom
	// handler for the 404 Not Found responses.
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	// Serve specific static file
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	// Create new middleware chain containing specific middleware for dynamic application.
	// Will only contain LoadAndSave middleware for now.
	dynamic := alice.New(app.sessionManager.LoadAndSave)

	/*
		mux.HandleFunc("/", app.home)
		mux.HandleFunc("/snippet/view", app.snippetView)
		mux.HandleFunc("/snippet/create", app.snippetCreate)
	*/
	// http method, pattern req url path must match, handler to dispatch to
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/snippet/create", dynamic.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", dynamic.ThenFunc(app.snippetCreatePost))

	// NOTE: logRequest ↔ secureHeaders ↔ servemux ↔ handler
	// return app.recoverPanic(app.logRequest(secureHeaders(mux)))

	// Create a middleware chain using the justinas/alice package
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Return the standard middleware chain followed by the servemux
	return standard.Then(router)
}
