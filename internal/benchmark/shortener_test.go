package benchmark

import (
	"context"
	"testing"

	"github.com/vladimirimekov/url-shortener/internal/handlers"
	"github.com/vladimirimekov/url-shortener/internal/storage"
)

type userIDtype string

const userKey userIDtype = "userid"

func BenchmarkShortener(b *testing.B) {

	const count = 10_000_000

	memoryVar := make(map[string]map[string]string)
	h := handlers.Handler{
		LengthOfShortname: 10,
		Host:              "http://localhost:8080",
		UserKey:           userKey,
		Storage:           storage.MemoryWork{UserData: memoryVar},
	}

	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		for k := 0; k < count; k++ {
			h.GetShortname(ctx)
		}
	}
}
