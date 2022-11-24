package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladimirimekov/url-shortener/internal/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_MainHandler(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name     string
		urlValue string
		wantPost want
		wantGet  want
	}{
		{
			name:     "sipmle url",
			urlValue: "https://google.com",
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated},
			wantGet:  want{contentType: "text/plain; charset=utf-8", statusCode: http.StatusTemporaryRedirect},
		},
		{
			name:     "long url",
			urlValue: "https://goiejrgoijergiojposd.com",
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated},
			wantGet:  want{contentType: "text/plain; charset=utf-8", statusCode: http.StatusTemporaryRedirect},
		},
		{
			name:     "long url with slugs",
			urlValue: "https://rthiiurgfougjfeorferguti.com/thgeufijrgeuhfjwer/gerhuiojgeuh",
			wantPost: want{contentType: "application/json", statusCode: http.StatusCreated},
			wantGet:  want{contentType: "text/plain; charset=utf-8", statusCode: http.StatusTemporaryRedirect},
		},
	}

	s := storage.Storage{"data.gob"}
	d := Handler{s, 8, "http://localhost:8080/"}

	for _, tt := range tests {
		var shortURL string
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := http.HandlerFunc(d.MainHandler)

			requestPost := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.urlValue))
			h.ServeHTTP(w, requestPost)
			resultPost := w.Result()

			assert.Equal(t, tt.wantPost.statusCode, resultPost.StatusCode)
			assert.Equal(t, tt.wantPost.contentType, resultPost.Header.Get("Content-Type"))

			shortname, err := io.ReadAll(resultPost.Body)
			require.NoError(t, err)
			err = resultPost.Body.Close()
			require.NoError(t, err)
			shortURL = string(shortname)
		})

		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := http.HandlerFunc(d.MainHandler)

			requestGet := httptest.NewRequest(http.MethodGet, shortURL, nil)
			h.ServeHTTP(w, requestGet)
			resultGet := w.Result()

			assert.Equal(t, tt.wantGet.statusCode, resultGet.StatusCode)
			assert.Equal(t, tt.wantGet.contentType, resultGet.Header.Get("Content-Type"))
			assert.Equal(t, tt.urlValue, resultGet.Header.Get("Location"))
		})

	}
}
