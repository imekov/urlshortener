package storage

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"os"
)

type FileSystemConnect struct {
	Filename string
}

func (s FileSystemConnect) openFile(flag int) *os.File {

	dataFile, err := os.OpenFile(s.Filename, flag|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return dataFile
}

func (s FileSystemConnect) ReadData() map[string]map[string]string {
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

func (s FileSystemConnect) SaveData(d map[string]map[string]string) error {
	data := s.ReadData()

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

func (s FileSystemConnect) DeleteData([]string, string) {
}

func (s FileSystemConnect) GetURLByShortname(shortname string) (string, bool) {

	data := s.ReadData()

	for _, value := range data {
		if originalURL, ok := value[shortname]; ok {
			return originalURL, false
		}
	}

	return "", false
}
