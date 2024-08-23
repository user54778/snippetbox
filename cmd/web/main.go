package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"snippetbox.adpollak.net/internal/models"
)

// This type holds application-wide dependencies for our webapp.
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	snippets *models.SnippetModel // NOTE: Make the SnippetModel available to our handlers.
}

// Wraps sql.Open() and returns a sql.DB connection pool for
// a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	// NOTE: sql.Open does NOT actually create any connections; just initializes the pool for FUTURE use.
	// Actual connections are establised lazily, we use Ping() to create connection and
	// verify things setup correctly.
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	// Define a cli arg named `addr`, w/ default value of :4000.
	// Additionally define some help text to explain flag controls
	addr := flag.String("addr", ":4000", "HTTP network address")
	// Define a new cli flag for the MySQL DSN string
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")

	// Parse CLI flag.
	// This reads in the CLI flag value and assigns it to addr.
	// NOTE: You MUST call this BEFORE you use the addr value, otherwise it
	// will ALWAYS use the default value.
	flag.Parse()

	// Create a new logger using log.New() for writing information messages.
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Create a new logger for writing error messages
	// log.Lshortfile flag includes relevant file name and line number.
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close() // NOTE:

	// Initialize a models.SnippetModel instance and add it to the application
	// dependencies.
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippets: &models.SnippetModel{DB: db},
	}
	/*
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
		mux.HandleFunc("/", app.home) // Refactored to be methods on the application type.
		mux.HandleFunc("/snippet/view", app.snippetView)
		mux.HandleFunc("/snippet/create", app.snippetCreate)
	*/

	// NOTE: Go's HTTP server by default uses the standard logger.
	// We initialize a new http.Server struct containing the config settings for
	// the server as opposed to using the ListenAndServe shortcut to use our custom error.
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(), // updated to call app.routes() to get the servemux containing our routes.
	}

	infoLog.Printf("Starting server on %s\n", *addr)
	// err := http.ListenAndServe(*addr, mux)
	err = srv.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			errorLog.Fatal(err)
		}
	}
}
