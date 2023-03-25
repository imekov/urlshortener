package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/lib/pq"
)

// PostgreConnect хранит соединение с базой данных.
type PostgreConnect struct {
	DBConnect *sql.DB
}

// URLRow используется для чтения данных из базы данных.
type URLRow struct {
	UserID      string
	ShortURL    string
	OriginalURL string
}

// GetNewConnection - конструктор PostgreConnect.
func GetNewConnection(db *sql.DB, dbConf string, migrationAddress string) PostgreConnect {

	dbConn := PostgreConnect{DBConnect: db}

	migration, err := migrate.New(migrationAddress, dbConf)
	if err != nil {
		log.Print(err)
	}

	if err = migration.Up(); errors.Is(err, migrate.ErrNoChange) {
		log.Print(err)
	}

	return dbConn
}

// ReadData читает данные из базы данных.
func (s PostgreConnect) ReadData(ctx context.Context) map[string]map[string]string {

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
	}
	defer tx.Rollback()

	data := map[string]map[string]string{}

	rows, err := s.DBConnect.Query("SELECT user_Cookie FROM users")
	if err != nil {
		log.Print(err)
	}

	defer rows.Close()

	for rows.Next() {
		var v string

		err = rows.Scan(&v)
		if err != nil {
			log.Print(err)
		}

		data[v] = map[string]string{}
	}

	err = rows.Err()
	if err != nil {
		log.Print(err)
	}

	sqlStatement := `
	SELECT 
    	users.user_Cookie,
		urls.shortURL,
		urls.originalURL	
	FROM
	urls
	INNER JOIN users ON users.user_ID = urls.user_ID;`

	rows, err = s.DBConnect.Query(sqlStatement)
	if err != nil {
		log.Print(err)
	}

	defer rows.Close()

	for rows.Next() {
		var v URLRow

		err = rows.Scan(&v.UserID, &v.ShortURL, &v.OriginalURL)
		if err != nil {
			log.Print(err)
		}

		data[v.UserID][v.ShortURL] = v.OriginalURL
	}

	err = rows.Err()
	if err != nil {
		log.Print(err)
	}

	tx.Commit()

	return data
}

// SaveData сохраняет данные в БД.
func (s PostgreConnect) SaveData(ctx context.Context, d map[string]map[string]string) error {

	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return err
	}
	defer tx.Rollback()

	sqlInsertUser, err := tx.Prepare("INSERT INTO users (user_Cookie) VALUES ($1) ON CONFLICT (user_Cookie) DO NOTHING;")
	if err != nil {
		log.Print(err)
		return err
	}

	defer sqlInsertUser.Close()

	sqlInsertData, err := tx.Prepare("INSERT INTO urls (user_ID, shortURL, originalURL) VALUES ((SELECT user_ID from users WHERE user_Cookie=$1), $2, $3);")
	if err != nil {
		log.Print(err)
		return err
	}

	defer sqlInsertData.Close()

	for userID, values := range d {

		_, err := sqlInsertUser.Exec(userID)
		if err != nil {
			log.Print(err)
			return err
		}

		for shortURL, originalURL := range values {

			_, err := sqlInsertData.Exec(userID, shortURL, originalURL)
			if err != nil {
				log.Print(err)
				return err
			}
		}
	}

	tx.Commit()
	return nil

}

// DeleteData удаляет данные из БД.
func (s PostgreConnect) DeleteData(data []string, user string) {
	tx, err := s.DBConnect.Begin()
	if err != nil {
		log.Print(err)
		return
	}
	defer tx.Rollback()

	sqlDeleteURLS, err := tx.Prepare("update urls set isDelete = true from (select unnest($1::text[]) as shortURL) as data_table where urls.shortURL = data_table.shortURL and urls.user_ID = (SELECT user_ID from users WHERE user_Cookie=$2);")

	if err != nil {
		log.Print(err)
		return
	}

	defer sqlDeleteURLS.Close()

	_, err = sqlDeleteURLS.Exec(pq.Array(data), user)
	if err != nil {
		log.Print(err)
		return
	}

	tx.Commit()
}

// GetURLByShortname возвращает из БД оригинальный URL на основе сокращенной ссылки.
func (s PostgreConnect) GetURLByShortname(ctx context.Context, shortname string) (originalURL string, isDelete bool) {
	tx, err := s.DBConnect.BeginTx(ctx, nil)
	if err != nil {
		log.Print(err)
		return "", false
	}
	defer tx.Rollback()

	err = tx.QueryRow("select originalURL, isDelete from urls where urls.shortURL = $1;", shortname).Scan(&originalURL, &isDelete)

	if err != nil {
		log.Print(err)
		return "", false
	}

	tx.Commit()
	return originalURL, isDelete
}

// PingDBConnection проверяет соединение с базой данных.
func (s PostgreConnect) PingDBConnection(ctx context.Context) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err := s.DBConnect.PingContext(ctxWithTimeout)
	return err

}
