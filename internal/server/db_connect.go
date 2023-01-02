package server

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func Connect(DBAddress string) *sql.DB {

	db, err := sql.Open("postgres", DBAddress)
	if err != nil {
		panic(err)
	}

	return db

}
