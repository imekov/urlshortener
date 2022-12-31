package server

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	port     = 5432
	user     = "postgres"
	password = "12345678"
	dbname   = "study_db"
)

func Connect(DBAddress string) *sql.DB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", DBAddress, port, user, password, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	return db

}
