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

	/*
		mux.HandleFunc("/", app.home)
		mux.HandleFunc("/snippet/view", app.snippetView)
		mux.HandleFunc("/snippet/create", app.snippetCreate)
	*/
	// http method, pattern req url path must match, handler to dispatch to
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/snippet/view/:id", app.snippetView)
	router.HandlerFunc(http.MethodGet, "/snippet/create", app.snippetCreate)
	router.HandlerFunc(http.MethodPost, "/snippet/create", app.snippetCreatePost)

	// Pass the servemux as the 'next' parameter to the secureHeaders middleware.
	// Added passing the servemux from secureHeaders to logRequest first
	// NOTE: logRequest ↔ secureHeaders ↔ servemux ↔ handler
	// return app.recoverPanic(app.logRequest(secureHeaders(mux)))

	// Create a middleware chain using the justinas/alice package
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// Return the standard middleware chain followed by the servemux
	return standard.Then(router)
}
