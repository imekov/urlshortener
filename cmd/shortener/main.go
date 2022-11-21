package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
)

const (
	filename          = "data.gob"
	lengthOfShortname = 8
)

func MainHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:
		data := ReadData()
		shortnameID := r.URL.Query().Get("id")

		if url, ok := data[shortnameID]; ok {
			w.Header().Set("content-type", "text/plain; charset=utf-8")
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "URL not found", 404)
		}

	case http.MethodPost:
		var shortname string
		savedData := ReadData()
		m := make(map[string]string)

		for {
			shortname = GenerateShortname(lengthOfShortname)
			if _, ok := savedData[shortname]; !ok {
				break
			}
		}

		bytesBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if err := json.Unmarshal(bytesBody, &m); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		for _, v := range m {
			savedData[shortname] = v
		}

		SaveData(savedData)

		subj := map[string]string{"shortURL": shortname}
		resp, err := json.Marshal(subj)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)

	default:
		http.Error(w, "Bad Request", 400)
	}
}

func GenerateShortname(length int) string {

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, length)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}

	return string(s)

}

func CheckFileExist() {

	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		emptyMap := make(map[string]string)
		dataFile, err := os.Create(filename)
		defer dataFile.Close()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		dataEncoder := gob.NewEncoder(dataFile)
		dataEncoder.Encode(emptyMap)
	}
}

func ReadData() map[string]string {
	var data map[string]string

	CheckFileExist()

	dataFile, err := os.Open(filename)
	defer dataFile.Close()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&data)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return data
}

func SaveData(d map[string]string) {

	dataFile, err := os.Create(filename)
	defer dataFile.Close()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(d)

}

func main() {

	http.HandleFunc("/", MainHandler)

	http.ListenAndServe(":8080", nil)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
