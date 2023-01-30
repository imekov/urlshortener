package storage

import (
	"errors"
	"net/http"
)

type MemoryWork struct {
	UserData map[string]map[string]string
}

func (s MemoryWork) ReadData() map[string]map[string]string {
	return s.UserData
}

func (s MemoryWork) SaveData(d map[string]map[string]string) error {

	for userID, values := range d {
		if _, ok := s.UserData[userID]; !ok {
			s.UserData[userID] = map[string]string{}
		}

		for shortURL, originalURL := range values {
			s.UserData[userID][shortURL] = originalURL
		}

	}

	return nil

}

func (s MemoryWork) DeleteData(arrayToDelete []string, user string) {

	for _, shortURL := range arrayToDelete {
		if _, isDelete := s.GetURLByShortname(shortURL); !isDelete {
			s.UserData[user][shortURL] = "-" + s.UserData[user][shortURL]
		}
	}
}

func (s MemoryWork) GetURLByShortname(shortname string) (string, bool) {

	for _, value := range s.UserData {
		if originalURL, ok := value[shortname]; ok {
			if originalURL[0] == '-' {
				return "", true
			}
			return originalURL, false
		}
	}

	return "", false
}

func (s MemoryWork) PingDBConnection(w http.ResponseWriter, r *http.Request) {
	err := errors.New("db is not working, current type - work with memory")
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
