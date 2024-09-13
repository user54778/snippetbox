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

	// NOTE: Unprotected application routes using the "dynamic" middleware chain
	dynamic := alice.New(app.sessionManager.LoadAndSave)

	// http method, pattern req url path must match, handler to dispatch to
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	// NOTE: protected (authenticated-only) application routes, use a new "protected"
	// middleware chain which includes the requireAuthentication middleware.
	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))

	// NOTE: logRequest ↔ secureHeaders ↔ servemux ↔ handler
	// return app.recoverPanic(app.logRequest(secureHeaders(mux)))
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	// Return the standard middleware chain followed by the servemux
	return standard.Then(router)
}
