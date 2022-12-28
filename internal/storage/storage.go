package storage

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

type Storage struct {
	Filename string
}

func (s Storage) ReadData() map[string]map[string]string {
	var data map[string]map[string]string

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
		data = map[string]map[string]string{}
	}

	return data
}

func (s Storage) SaveData(d map[string]map[string]string) {

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

func (s Storage) Encrypt(stringToEncrypt string, keyString string) (encryptedString string, err error) {

	plaintext := []byte(stringToEncrypt)

	block, err := aes.NewCipher([]byte(keyString))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	return hex.EncodeToString(ciphertext), nil
}

func (s Storage) Decrypt(encryptedString string, keyString string) (decryptedString string, err error) {

	enc, err := hex.DecodeString(encryptedString)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(keyString))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(enc) < nonceSize {
		return "", errors.New("length of encrypted string is too short")
	}

	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
