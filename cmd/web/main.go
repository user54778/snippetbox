package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"snippetbox.adpollak.net/internal/models"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

// This type holds application-wide dependencies for our webapp.
type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       *models.SnippetModel          // NOTE: Make the SnippetModel available to our handlers.
	users          *models.UserModel             // Same as SnippetModel
	templateCache  map[string]*template.Template // make avail cache to our handlers
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
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

	// Initialize new template cache
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	// NOTE: Initialize a new sessionManager. Configured to use
	// our MySQL db as the session store, and set a lifetime of 12 hours.
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	// Initialize a models.SnippetModel instance and add it to the application
	// dependencies.
	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// Initialize a tlsConfig struct to hold non-default TLS settings we want the server to use.
	// Only changing curve preference values, so only elliptic curves with assemply impls
	// will be used.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// NOTE: Go's HTTP server by default uses the standard logger.
	// We initialize a new http.Server struct containing the config settings for
	// the server as opposed to using the ListenAndServe shortcut to use our custom error.
	srv := &http.Server{
		Addr:      *addr,
		ErrorLog:  errorLog,
		Handler:   app.routes(), // updated to call app.routes() to get the servemux containing our routes.
		TLSConfig: tlsConfig,
		// Add Idle, Read, and Write timeouts to the server
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s\n", *addr)
	// err := http.ListenAndServe(*addr, mux)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}
