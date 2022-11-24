package main

import (
	"github.com/vladimirimekov/url-shortener/internal/handlers"
	"github.com/vladimirimekov/url-shortener/internal/storage"
	"log"
	"net/http"
)

const (
	filename          = "data.gob"
	lengthOfShortname = 8
	hostname          = "http://localhost:8080/"
)

func main() {

	s := storage.Storage{Filename: filename}
	h := handlers.Handler{Storage: s, LengthOfShortname: lengthOfShortname, Host: hostname}

	http.HandleFunc("/", h.MainHandler)

	http.ListenAndServe(":8080", nil)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
