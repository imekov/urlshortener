package main

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"

	_ "net/http/pprof"

	_ "github.com/lib/pq"
	"github.com/vladimirimekov/url-shortener/internal/server"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)
	var dbConnection *sql.DB
	defer dbConnection.Close()

	go func() {
		http.ListenAndServe("127.0.0.1:9999", nil)
	}()

	cfg, router := server.GetServer(dbConnection)

	if cfg.EnableHttps {
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("localhost", "127.0.0.1"),
		}

		httpsServer := &http.Server{
			Addr:      ":443",
			Handler:   router,
			TLSConfig: manager.TLSConfig(),
		}
		log.Fatal(httpsServer.ListenAndServeTLS("", ""))
	} else {
		log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
	}

}
