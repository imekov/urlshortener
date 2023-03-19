package main

import (
	"database/sql"
	"fmt"
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

	log.Fatal(http.ListenAndServe(server.GetServer(dbConnection)))
}
