package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	createDb   = flag.Bool("n", false, "create new, empty database")
	runTLS     = flag.Bool("t", false, "run with TLS")
	portNumber = flag.String("p", "8080", "port number to listen on")
	lock       = sync.Mutex{}
)

func main() {
	flag.Parse()
	var db *sql.DB
	if *createDb {
		var err error
		os.Remove("./ecc.db")
		db, err = sql.Open("sqlite3", "./ecc.db")
		if err != nil {
			log.Fatal(err)
		}

		sqlStmt := `
	create table codes (path text not null primary key, password text, valueMap text);
	delete from codes;
	`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Fatal(err)
		}

		// insert dummy data
		createNewCode(db, "dumdum", "password123")
	} else {
		var err error
		db, err = sql.Open("sqlite3", "./ecc.db")
		if err != nil {
			log.Fatal(err)
		}
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RedirectSlashes)

	r.Get("/", getIndex(db))
	r.Get("/{id}", getCode(db))
	r.Post("/{id}/encode", postEncode(db))
	r.Get("/{id}/encode", getCode(db))
	r.Post("/{id}/decode", postDecode(db))
	r.Get("/{id}/decode", getCode(db))
	r.Post("/{id}/save", postSaveMap(db))
	r.Get("/{id}/save", getCode(db))

	// Start server
	if *runTLS {
		autoTLSManager := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache(".cache"),
			//HostPolicy: autocert.HostWhitelist("<DOMAIN>"),
		}
		s := http.Server{
			Addr:    ":443",
			Handler: r,
			TLSConfig: &tls.Config{
				GetCertificate: autoTLSManager.GetCertificate,
				NextProtos:     []string{acme.ALPNProto},
			},
		}
		if err := s.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	} else {
		if err := http.ListenAndServe(":"+*portNumber, r); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}
}
