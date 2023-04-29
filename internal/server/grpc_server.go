package server

import (
	"google.golang.org/grpc"

	"github.com/vladimirimekov/url-shortener/internal/handlers"
	pb "github.com/vladimirimekov/url-shortener/proto"
)

// NewGRPCServer создает новый rpc сервис
func NewGRPCServer(handler handlers.Handler) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterUrlShortenerServer(s, handler)
	return s
}
