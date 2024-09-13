package main

type contextKey string

// Unique key to use to store and retrieve authentication status from a
// request context.
const isAuthenticatedContextKey = contextKey("isAuthenticated")
