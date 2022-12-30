package middlewares

import (
	"context"
	"github.com/google/uuid"
	"github.com/vladimirimekov/url-shortener/internal/server"
	"net/http"
)

type Repositories interface {
	ReadData() map[string]map[string]string
	SaveData(map[string]map[string]string)
}

type UserCookies struct {
	Storage Repositories
	Secret  string
	UserKey interface{}
}

func (h UserCookies) CheckUserCookies(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		st, err := r.Cookie("session_token")

		if err == nil {
			userID, errDecrypt := server.Decrypt(st.Value, h.Secret)
			savedData := h.Storage.ReadData()
			_, ok := savedData[userID]

			if errDecrypt == nil && ok {
				ctx := context.WithValue(r.Context(), h.UserKey, userID)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
		}

		sessionToken := uuid.NewString()
		savedData := h.Storage.ReadData()
		savedData[sessionToken] = map[string]string{}
		h.Storage.SaveData(savedData)

		enc, err := server.Encrypt(sessionToken, h.Secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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
