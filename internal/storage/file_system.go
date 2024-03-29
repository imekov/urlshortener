package storage

import (
	"bufio"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
)

// FileSystemConnect содержит имя файла.
type FileSystemConnect struct {
	Filename string
}

// openFile возвращает ссылку на открытый файл.
func (s FileSystemConnect) openFile(flag int) *os.File {

	dataFile, err := os.OpenFile(s.Filename, flag|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return dataFile
}

// ReadData читает данные из файла.
func (s FileSystemConnect) ReadData(context.Context) map[string]map[string]string {
	var data map[string]map[string]string

	dataFile := s.openFile(os.O_RDONLY)

	defer func(File *os.File) {
		err := File.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}(dataFile)

	err := gob.NewDecoder(dataFile).Decode(&data)

	if err != nil {
		data = map[string]map[string]string{}
	}

	return data
}

// SaveData сохраняет данные в файл.
func (s FileSystemConnect) SaveData(ctx context.Context, d map[string]map[string]string) error {
	data := s.ReadData(ctx)

	for userID, values := range d {
		if _, ok := data[userID]; !ok {
			data[userID] = map[string]string{}
		}

		for shortURL, originalURL := range values {
			data[userID][shortURL] = originalURL
		}

	}

	dataFile := s.openFile(os.O_WRONLY)
	defer dataFile.Close()

	writer := bufio.NewWriter(dataFile)

	err := gob.NewEncoder(writer).Encode(&data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil

}

// DeleteData удаляет данные из файла.
func (s FileSystemConnect) DeleteData(arrayToDelete []string, user string) {
	data := s.ReadData(context.Background())

	for _, shortURL := range arrayToDelete {
		if _, isDelete := s.GetURLByShortname(context.Background(), shortURL); !isDelete {
			data[user][shortURL] = "-" + data[user][shortURL]
		}
	}

	err := s.SaveData(context.Background(), data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// GetURLByShortname возвращает исходный URL на основе исходной ссылки.
func (s FileSystemConnect) GetURLByShortname(ctx context.Context, shortname string) (string, bool) {

	data := s.ReadData(ctx)

	for _, value := range data {
		if originalURL, ok := value[shortname]; ok {
			if originalURL[0] == '-' {
				return "", true
			}
			return originalURL, false
		}
	}

	return "", false
}

// PingDBConnection - заглушка для интерфейса.
func (s FileSystemConnect) PingDBConnection(ctx context.Context) error {
	err := errors.New("db is not working, current type - work with files")
	return err
}

// GetStatistic - возвращает количество ссылок и пользователей
func (s FileSystemConnect) GetStatistic() (urls int, users int) {
	data := s.ReadData(context.Background())
	for _, v := range data {
		urls += len(v)
	}

	return urls, len(data)
}
