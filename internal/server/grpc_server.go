package server

import (
	"github.com/vladimirimekov/url-shortener/internal/handlers"
	pb "github.com/vladimirimekov/url-shortener/proto"
	"google.golang.org/grpc"
)

// NewGRPCServer создает новый rpc сервис
func NewGRPCServer(baseURL string, handler handlers.Handler) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterUrlShortenerServer(s, handler)
	return s
}
