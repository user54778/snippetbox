package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"snippetbox.adpollak.net/ui"
)

// Returns a ServeMux containing our application routes.
// NOTE: we changed the return type from ServeMux to Handler
// since we wanted to wrap our middleware around the ServeMux.
func (app *application) routes() http.Handler {
	// Initialize the router
	router := httprouter.New()

	// Create a handler function that wraps our notFound helper, and then assigns it as the custom
	// handler for the 404 Not Found responses.
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	// Take the ui.Files embedded filesystem and convert it to a http.FS type
	// such that it satisfies the http.FileSystem interface. We then
	// pass that to the http.FileServer() function to create the file server handler.
	fileServer := http.FileServer(http.FS(ui.Files))

	// Serve specific static file
	// Static files contained in static folder of ui.Files embedded filesystem.
	// So, we no longer need to strip the prefix from the request url, since any requests now
	// starting with /static/ can just be passed directly to the file server and the corresponding
	// static file will be served (as long as it exists).
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	// Add a new GET /ping route, which calls HandlerFunc since it's not a part of any
	// middleware chain.
	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// NOTE: Unprotected application routes using the "dynamic" middleware chain
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// http method, pattern req url path must match, handler to dispatch to
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	// About route
	router.Handler(http.MethodGet, "/about", dynamic.ThenFunc(app.about))
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
	router.Handler(http.MethodGet, "/account/view", protected.ThenFunc(app.accountView))
	router.Handler(http.MethodGet, "/account/password/update", protected.ThenFunc(app.accountPasswordUpdate))
	router.Handler(http.MethodPost, "/account/password/update", protected.ThenFunc(app.accountPasswordUpdatePost))

	// NOTE: logRequest ↔ secureHeaders ↔ servemux ↔ handler
	// return app.recoverPanic(app.logRequest(secureHeaders(mux)))
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	// Return the standard middleware chain followed by the servemux
	return standard.Then(router)
}
