package server

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"net"
)

const (
	user     = "postgres"
	password = "12345678"
	dbname   = "study_db"
)

func Connect(DBAddress string) *sql.DB {
	host, port, err := net.SplitHostPort(DBAddress)
	if err != nil {
		panic(err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	return db

}
