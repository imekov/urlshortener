package middlewares

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type Repositories interface {
	ReadData() map[string]map[string]string
	SaveData(map[string]map[string]string) error
}

type UserCookies struct {
	Storage Repositories
	Secret  []byte
	UserKey interface{}
}

func (h UserCookies) CheckUserCookies(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		st, err := r.Cookie("session_token")

		if err == nil {

			var errorDecode error

			enc, err := hex.DecodeString(st.Value)
			if err != nil {
				errorDecode = err
			}

			block, err := aes.NewCipher(h.Secret)
			if err != nil {
				errorDecode = err
			}

			aesGCM, err := cipher.NewGCM(block)
			if err != nil {
				errorDecode = err
			}

			nonceSize := aesGCM.NonceSize()
			if len(enc) < nonceSize {
				errorDecode = err
			}

			nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

			plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				errorDecode = err
			}

			userID := string(plaintext)

			savedData := h.Storage.ReadData()
			if _, ok := savedData[userID]; errorDecode == nil && ok {
				ctx := context.WithValue(r.Context(), h.UserKey, userID)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
		}

		sessionToken := uuid.NewString()
		savedData := make(map[string]map[string]string)
		savedData[sessionToken] = map[string]string{}
		h.Storage.SaveData(savedData)

		plaintext := []byte(sessionToken)

		block, err := aes.NewCipher(h.Secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		aesGCM, err := cipher.NewGCM(block)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nonce := make([]byte, aesGCM.NonceSize())
		if _, err = io.ReadFull(cryptoRand.Reader, nonce); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

		enc := hex.EncodeToString(ciphertext)

		ctx := context.WithValue(r.Context(), h.UserKey, sessionToken)
		r = r.WithContext(ctx)

		http.SetCookie(w, &http.Cookie{
			Name:  "session_token",
			Value: enc,
			Path:  "/",
		})
		next.ServeHTTP(w, r)
	})

}
