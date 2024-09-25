package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)

// This is a middleware function.
// It accepts the next handler in a chain of ServeHTTP() methods as a parameter.
// It returns a handler that executes the setting of security headers and
// then calls the next handler as a return over an anonymous function.
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")

		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

// This middleware is a method against application that also has access to
// the handler dependencies including the information logger (which we need).
// This middleware method logs HTTP request.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

// A middleware method to recover from a pnaic and call our serverError() helper
// to give a better user response when a goroutine panics.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always runs in the event of a panic as Go unwinds the call stack.
		defer func() {
			// Use recover() to check if panic has occurred.
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")

				// Call serverError to return a 500 Internal Server response.
				// Normalize the any param from recover() by using fmt.Errorf to create a new
				// error object containing default text representation of an any value.
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// NOTE: authentication/authorization/security middleware

// Middleware to prevent unauthenticated users from attempting to visit
// any routes with URL path /snippet/create or any other that requires
// to be logged in.
func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			// Add path user try to access to session data.
			app.sessionManager.Put(r.Context(), "redirectPathAfterLogin", r.URL.Path)
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Set the Cache-Control: no-store header so that pages that require
		// authentication are not stored in the users browser cache (or other
		// intermediary cache).
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1) Retrieve the user's ID from the session data.
		// We use GetInt() to accomplish this; it returns a zero value (for int that is 0)
		// if no "authenticatedUserID" value is in the session.
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}

		// 2) Check the database to see if a user with that ID corresponds to
		// a valid user.
		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(w, err)
			return
		}

		// 3) Matching user found; Update the request context to include
		// an isAuthenticatedContextKey with the value true.
		// Create a new copy of the request ctx and assign to r.
		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// Middleware that uses a customized CSRF cookie with
// Secure, Path, and HttpOnly attributes set.
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}
