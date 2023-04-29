package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	pb "github.com/vladimirimekov/url-shortener/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

// Repositories - интерфейс с методами для работы модуля.
type Repositories interface {
	ReadData(context.Context) map[string]map[string]string
	SaveData(context.Context, map[string]map[string]string) error
	DeleteData([]string, string)
	GetURLByShortname(context.Context, string) (string, bool)
	PingDBConnection(ctx context.Context) error
	GetStatistic() (int, int)
}

// Handler хранит базовые настройки хэндлера и интерфейс с методами для работы с хэнделами.
type Handler struct {
	Storage           Repositories
	LengthOfShortname int
	Host              string
	UserKey           interface{}
	pb.UnimplementedUrlShortenerServer
}

// GetData содержит структуру для получения ссылок в формате json.
type GetData struct {
	URL string `json:"url"`
}

// AllUserURLs содержит структуру для json данных со всеми пользовательскими URL.
type AllUserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// Statistic содержит структуру для json данных со статистикой.
type Statistic struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}

// BatchData содержит структуру для получения json данных с пачкой ссылок для сокращения.
type BatchData struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url,omitempty"`
	ShortURL      string `json:"short_url"`
}

// GetShortname возвращает неиспользуемую раннее строку для сокращения ссылок.
func (h Handler) GetShortname(ctx context.Context) string {
	var shortname string
	var result bool

	savedData := h.Storage.ReadData(ctx)

	//проверка на существование сгенерированного имени
	for {
		result = true

		letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

		s := make([]rune, h.LengthOfShortname)
		for i := range s {
			s[i] = letters[rand.Intn(len(letters))]
		}

		shortname = string(s)

		for _, value := range savedData {
			if _, ok := value[shortname]; ok {
				result = false
				break
			}
		}

		if result {
			break
		}
	}

	return shortname
}

// Внутренняя функция для получения айди из контекста
func (h Handler) getUserID(r *http.Request) (string, error) {
	var userID string
	uk := r.Context().Value(h.UserKey)

	switch v := uk.(type) {
	default:
		err := fmt.Errorf("unexpected type of user key %T", v)
		return "", err
	case []byte:
		userID = string(uk.([]byte))
	case string:
		userID = uk.(string)
	}

	return userID, nil
}

// MainHandler содержит 2 в основных метода для получения сокращенной ссылки через POST запрос и получение оригинальной
// ссылки через GET запрос.
func (h Handler) MainHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodGet:

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		r = r.WithContext(ctx)

		shortname := chi.URLParam(r, "id")
		w.Header().Set("content-type", "text/plain; charset=utf-8")

		if originalURL, isDelete := h.Storage.GetURLByShortname(ctx, shortname); isDelete {
			w.WriteHeader(http.StatusGone)
		} else if originalURL == "" {
			http.Error(w, "URL not found", http.StatusNotFound)
		} else {
			w.Header().Set("Location", originalURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		}

	case http.MethodPost:

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		r = r.WithContext(ctx)

		userID, err := h.getUserID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}(r.Body)

		currentURL := string(b)

		_, err = url.ParseRequestURI(currentURL) //проверка на валидность url
		if err != nil {
			http.Error(w, "Invalid URL value", http.StatusBadRequest)
			return
		}

		shortname := h.GetShortname(ctx)
		resultData := map[string]map[string]string{userID: {shortname: currentURL}}

		if err = h.Storage.SaveData(ctx, resultData); err != nil {
			switch e := err.(type) {
			case *pq.Error:
				if pgerrcode.IsIntegrityConstraintViolation(string(e.Code)) {
					w.Header().Set("content-type", "application/json")
					w.WriteHeader(http.StatusConflict)

					savedData := h.Storage.ReadData(ctx)

					for _, value := range savedData {
						for short, original := range value {
							if original == currentURL {
								_, err = w.Write([]byte(h.Host + "/" + short))
								if err != nil {
									http.Error(w, err.Error(), http.StatusInternalServerError)
									return
								}
								break
							}
						}
					}

					return
				}
			default:
				http.Error(w, e.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(h.Host + "/" + shortname))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "Bad Request", http.StatusBadRequest)

	}
}

// PostShortenHandler создает сокращенную ссылку из URL, полученного в json формате.
func (h Handler) PostShortenHandler(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	userID, err := h.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}(r.Body)

	g := GetData{}

	if err := json.Unmarshal(b, &g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//проверка на валидность url
	_, err = url.ParseRequestURI(g.URL)
	if err != nil {
		http.Error(w, "Invalid URL value", http.StatusBadRequest)
		return
	}

	shortname := h.GetShortname(ctx)

	resultData := map[string]map[string]string{userID: {shortname: g.URL}}

	if err = h.Storage.SaveData(ctx, resultData); err != nil {
		switch e := err.(type) {
		case *pq.Error:
			if pgerrcode.IsIntegrityConstraintViolation(string(e.Code)) {
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusConflict)

				savedData := h.Storage.ReadData(ctx)

				for _, value := range savedData {
					for short, original := range value {
						if original == g.URL {

							resultJSON, err := json.Marshal(map[string]string{"result": h.Host + "/" + short})

							if err != nil {
								http.Error(w, err.Error(), http.StatusInternalServerError)
								return
							}

							_, err = w.Write(resultJSON)
							if err != nil {
								http.Error(w, err.Error(), http.StatusInternalServerError)
								return
							}

							break
						}
					}
				}

				return
			}
		default:
			http.Error(w, e.Error(), http.StatusInternalServerError)
			return
		}
	}

	resultJSON, err := json.Marshal(map[string]string{"result": h.Host + "/" + shortname})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resultJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// PostShortenBatchHandler получает ссылки списком, после чего возвращает итоговый массив с сокращенными URL.
func (h Handler) PostShortenBatchHandler(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	userID, err := h.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}(r.Body)

	var g []BatchData

	if err := json.Unmarshal(b, &g); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataToSave := make(map[string]map[string]string)
	dataToSave[userID] = map[string]string{}

	for index, value := range g {
		shortname := h.GetShortname(ctx)
		dataToSave[userID][shortname] = value.OriginalURL
		g[index].ShortURL = h.Host + "/" + shortname
		g[index].OriginalURL = ""
	}

	if err = h.Storage.SaveData(ctx, dataToSave); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resultJSON, err := json.Marshal(g)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resultJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetAllShorterURLsHandler отправляет список всех сокращенных ссылок текущего пользователя.
func (h Handler) GetAllShorterURLsHandler(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	userID, err := h.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	savedData := h.Storage.ReadData(ctx)
	userData := savedData[userID]

	if len(userData) == 0 {
		err := errors.New("there are no shortened links")
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	result := make([]AllUserURLs, 0, len(userData))

	for key, value := range userData {
		result = append(result, AllUserURLs{ShortURL: h.Host + "/" + key, OriginalURL: value})
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resultJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteBatchURLS выполняет асинхронное удаление ссылок.
func (h Handler) DeleteBatchURLS(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	userID, err := h.getUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}(r.Body)

	var s []string

	if err = json.Unmarshal(b, &s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	g, ctx := errgroup.WithContext(r.Context())

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				err := errors.New("context canceled")
				return err
			default:
				h.Storage.DeleteData(s, userID)
				return nil
			}
		}
	})

	if err = g.Wait(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// PingConnection проверяет соединение с базой данных.
func (h Handler) PingConnection(w http.ResponseWriter, r *http.Request) {
	if err := h.Storage.PingDBConnection(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetStatistics возвращает количество сокращённых URL в сервисе и количество пользователей в сервисе.
func (h Handler) GetStatistics(w http.ResponseWriter, r *http.Request) {

	urls, users := h.Storage.GetStatistic()
	result := Statistic{Urls: urls, Users: users}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resultJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateShortLink создает короткую ссылку для grpc.
func (h Handler) CreateShortLink(ctx context.Context, request *pb.CreateShortLinkRequest) (*pb.CreateShortLinkResponse, error) {
	var response pb.CreateShortLinkResponse

	shortname := h.GetShortname(ctx)
	resultData := map[string]map[string]string{request.UserID: {shortname: request.OriginalURL}}

	if err := h.Storage.SaveData(ctx, resultData); err != nil {
		switch e := err.(type) {
		case *pq.Error:
			if pgerrcode.IsIntegrityConstraintViolation(string(e.Code)) {
				savedData := h.Storage.ReadData(ctx)

				for _, value := range savedData {
					for short, original := range value {
						if original == request.OriginalURL {
							response.ShortURL = short
							break
						}
					}
				}

				return &response, nil
			}
		default:
			return nil, err
		}
	}

	response.ShortURL = h.Host + "/" + shortname

	return &response, nil
}

// GetOriginalLink возвращает короткую ссылку для grpc .
func (h Handler) GetOriginalLink(ctx context.Context, request *pb.GetOriginalLinkRequest) (*pb.GetOriginalLinkResponse, error) {
	var response pb.GetOriginalLinkResponse

	if originalURL, isDelete := h.Storage.GetURLByShortname(ctx, request.ShortURL); isDelete {
		return nil, errors.New("this link has been removed")
	} else if originalURL == "" {
		return nil, errors.New("URL not found")
	} else {
		response.OriginalURL = originalURL
	}

	return &response, nil
}

// CreateLinksInBatches создает батч ссылок.
func (h Handler) CreateLinksInBatches(ctx context.Context, request *pb.CreateLinksInBatchesRequest) (*pb.CreateLinksInBatchesResponse, error) {

	var response pb.CreateLinksInBatchesResponse

	dataToSave := make(map[string]map[string]string)
	dataToSave[request.UserID] = map[string]string{}

	for _, value := range request.OriginalURLs {
		shortname := h.GetShortname(ctx)
		dataToSave[request.UserID][shortname] = value.OriginalURL
		response.ShortURLs = append(response.ShortURLs, &pb.BatchResponse{ShortURL: h.Host + "/" + shortname, CorrelationID: value.CorrelationID})
	}

	if err := h.Storage.SaveData(ctx, dataToSave); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetAllShorterURLs возвращает все ссылки пользователя.
func (h Handler) GetAllShorterURLs(ctx context.Context, request *pb.GetAllShorterURLsRequest) (*pb.GetAllShorterURLsResponse, error) {
	var response pb.GetAllShorterURLsResponse

	savedData := h.Storage.ReadData(ctx)
	userData := savedData[request.UserID]

	if len(userData) == 0 {
		return nil, errors.New("there are no shortened links")
	}

	for key, value := range userData {
		response.ShortURLs = append(response.ShortURLs, &pb.AllShorterURLsResponse{ShortURL: h.Host + "/" + key, OriginalURL: value})
	}

	return &response, nil
}

// DeleteURLS удаляет.
func (h Handler) DeleteURLS(ctx context.Context, request *pb.DeleteURLSRequest) (*emptypb.Empty, error) {

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return errors.New("context canceled")
			default:
				h.Storage.DeleteData(request.ShortURLs, request.UserID)
				return nil
			}
		}
	})

	if err := g.Wait(); err != nil {
		return &emptypb.Empty{}, err
	}

	return &emptypb.Empty{}, nil
}

// PingDBConnection пингует.
func (h Handler) PingDBConnection(ctx context.Context, _ *emptypb.Empty) (*pb.PingDBConnectionResponse, error) {
	if err := h.Storage.PingDBConnection(ctx); err != nil {
		return nil, err
	}

	return &pb.PingDBConnectionResponse{Ok: true}, nil
}

// GetStats возвращает стату.
func (h Handler) GetStats(_ context.Context, _ *emptypb.Empty) (*pb.GetStatsResponse, error) {

	urls, users := h.Storage.GetStatistic()
	result := &pb.GetStatsResponse{Urls: int64(urls), Users: int64(users)}

	return result, nil
}
