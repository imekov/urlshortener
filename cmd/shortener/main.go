package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/acme/autocert"

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

	if cfg.EnableHTTPS {
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

		var srv = &http.Server{
			Addr:    cfg.ServerAddress,
			Handler: router,
		}

		idleConnsClosed := make(chan struct{})
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

		go func() {
			<-sigint
			if err := srv.Shutdown(context.Background()); err != nil {
				log.Printf("HTTP server Shutdown: %v", err)
			}
			close(idleConnsClosed)
		}()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
		<-idleConnsClosed
		os.Exit(0)
	}

}
