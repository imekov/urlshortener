package storage

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/lib/pq"
)

type PostgreConnect struct {
	DBConnect *sql.DB
}

type URLRow struct {
	UserID      string
	ShortURL    string
	OriginalURL string
}

func (s PostgreConnect) ReadData() map[string]map[string]string {
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

	return data
}

func (s PostgreConnect) SaveData(d map[string]map[string]string) error {

	tx, err := s.DBConnect.Begin()
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

func (s PostgreConnect) GetURLByShortname(shortname string) (originalURL string, isDelete bool) {
	tx, err := s.DBConnect.Begin()
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

func (s PostgreConnect) PingDBConnection(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := s.DBConnect.PingContext(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
