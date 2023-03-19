package storage

import (
	"context"
	"database/sql"
	"github.com/vladimirimekov/url-shortener/internal"
	"testing"

	_ "github.com/lib/pq"
)

func TestPostgreStorage_WriteReadData(t *testing.T) {
	cfg := internal.GetConfig()

	tests := []struct {
		name string
		want map[string]map[string]string
	}{
		{"string numbers", map[string]map[string]string{"gehfuii": {"12345": "6543"}}},
		{"long string", map[string]map[string]string{"hgudfjsi": {"hgutrhgitrhgoiwejoirjwoeijgeiojgoierg": "oigrjtohijroithjoirtjhoirtjhoirtjhoirjtioh"}}},
		{"mix", map[string]map[string]string{"hitrojg": {"8394ht98ghrfjuidrjf8943u": "65gi43hhfr&^#Grh2"}}},
		{"cyrillic", map[string]map[string]string{"jtyhrgef": {"проверка": "связи"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dbConnection, _ := sql.Open("postgres", cfg.DBAddress)
			defer dbConnection.Close()
			s := GetNewConnection(dbConnection, cfg.DBAddress, "file://../../migrations/postgres")
			ctx := context.Background()

			s.SaveData(ctx, tt.want)
			s.ReadData(ctx)

			for key, value := range tt.want {
				for k := range value {
					s.GetURLByShortname(ctx, k)
					s.DeleteData([]string{k}, key)
				}
			}

			s.PingDBConnection(ctx)

		})

	}
}
