package storage

import (
	"encoding/gob"
	"fmt"
	"os"
)

type Storage struct {
	Filename string
}

func (s Storage) ReadData() map[string]string {
	var data map[string]string

	_, err := os.Stat(s.Filename)
	if os.IsNotExist(err) {
		s.SaveData(map[string]string{})
	}

	dataFile, err := os.Open(s.Filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer dataFile.Close()

	dataDecoder := gob.NewDecoder(dataFile)
	err = dataDecoder.Decode(&data)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return data
}

func (s Storage) SaveData(d map[string]string) {

	dataFile, err := os.Create(s.Filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer dataFile.Close()

	dataEncoder := gob.NewEncoder(dataFile)
	dataEncoder.Encode(d)

}
