package storage

import (
	"context"
	"errors"
)

// MemoryWork хранит данные в мапе.
type MemoryWork struct {
	UserData map[string]map[string]string
}

// ReadData возвращает мапу с данными.
func (s MemoryWork) ReadData(context.Context) map[string]map[string]string {
	return s.UserData
}

// SaveData сохраняет пользовательские ссылки в память.
func (s MemoryWork) SaveData(_ context.Context, d map[string]map[string]string) error {

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

// DeleteData помечает на удаление сохранённые ссылки.
func (s MemoryWork) DeleteData(arrayToDelete []string, user string) {

	for _, shortURL := range arrayToDelete {
		if _, isDelete := s.GetURLByShortname(context.Background(), shortURL); !isDelete {
			s.UserData[user][shortURL] = "-" + s.UserData[user][shortURL]
		}
	}
}

// GetURLByShortname возвращает оригинальный URL из памяти на основе сокращённок ссылки.
func (s MemoryWork) GetURLByShortname(_ context.Context, shortname string) (string, bool) {

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

// PingDBConnection - заглушка для работы интерфейса.
func (s MemoryWork) PingDBConnection(context.Context) error {
	err := errors.New("db is not working, current type - work with memory")
	return err
}

// GetStatistic возвращает данные статистики
func (s MemoryWork) GetStatistic() (urls int, users int) {
	for _, v := range s.UserData {
		urls += len(v)
	}

	return urls, len(s.UserData)
}
