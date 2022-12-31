package server

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func Connect(DBAddress string) *sql.DB {
	db, err := sql.Open("sqlite3", DBAddress)
	if err != nil {
		panic(err)
	}

	return db

}
