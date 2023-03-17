package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "net/http/pprof"

	_ "github.com/lib/pq"
	"github.com/vladimirimekov/url-shortener/internal/server"
)

func main() {

	var dbConnection *sql.DB
	defer dbConnection.Close()

	go func() {
		http.ListenAndServe("127.0.0.1:9999", nil)
	}()

	log.Fatal(http.ListenAndServe(server.GetServer(dbConnection)))
}
