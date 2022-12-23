package storage

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"os"
)

type Storage struct {
	Filename string
}

func (s Storage) ReadData() map[string]string {
	var data map[string]string

	dataFile, err := os.OpenFile(s.Filename, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func(dataFile *os.File) {
		err := dataFile.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}(dataFile)

	err = gob.NewDecoder(dataFile).Decode(&data)

	if err != nil {
		data = map[string]string{}
	}

	return data
}

func (s Storage) SaveData(d map[string]string) {

	dataFile, err := os.OpenFile(s.Filename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func(dataFile *os.File) {
		err := dataFile.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}(dataFile)

	writer := bufio.NewWriter(dataFile)

	err = gob.NewEncoder(writer).Encode(&d)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
