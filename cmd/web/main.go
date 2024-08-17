package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	// Define a cli arg named `addr`, w/ default value of :4000.
	// Additionally define some help text to explain flag controls
	addr := flag.String("addr", ":4000", "HTTP network address")

	// Parse CLI flag.
	// This reads in the CLI flag value and assigns it to addr.
	// NOTE: You MUST call this BEFORE you use the addr value, otherwise it
	// will ALWAYS use the default value.
	flag.Parse()

	// Initialize a new servemux, then register the `home` function
	// as the handler for the "/" pattern.
	mux := http.NewServeMux()

	// Create a file server which SERVES files out of the "./ui/static" directory.
	// The path given is relative
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Register the file server as a handler for all URL paths that start with /static/.
	// NOTE: we must strip the /static prefix prior to the request receiving the server.
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// NOTE: HandleFunc is described as so: type Handler interface {
	//                                        ServeHTTP(http.ResponseWriter, *http.Request)
	//                                      }
	// The pointer indicates you may modify the request, however, most handlers will just read from it.
	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)

	log.Printf("Starting server on %s\n", *addr)
	err := http.ListenAndServe(*addr, mux)
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}
