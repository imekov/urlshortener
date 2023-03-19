package storage

import (
	"context"
	"log"
	"os"
	"reflect"
	"testing"
)

// TestStorage_WriteReadData тестирует сохранение данных в файл.
func TestStorage_WriteReadData(t *testing.T) {
	tests := []struct {
		name string
		file string
		want map[string]map[string]string
	}{
		{"string numbers", "tg3948he.gob", map[string]map[string]string{"gehfuii": {"12345": "6543"}}},
		{"long string", "1Jo$@gid%fg.gob", map[string]map[string]string{"hgudfjsi": {"hgutrhgitrhgoiwejoirjwoeijgeiojgoierg": "oigrjtohijroithjoirtjhoirtjhoirtjhoirjtioh"}}},
		{"mix", "32_489.gob", map[string]map[string]string{"hitrojg": {"8394ht98ghrfjuidrjf8943u": "65gi43hhfr&^#Grh2"}}},
		{"cyrillic", "hj4589gerio.gob", map[string]map[string]string{"jtyhrgef": {"проверка": "связи"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := FileSystemConnect{
				Filename: tt.file,
			}
			ctx := context.Background()
			s.SaveData(ctx, tt.want)
			if got := s.ReadData(ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadData() = %v, want %v", got, tt.want)
			}

			for key, value := range tt.want {
				for k := range value {
					s.GetURLByShortname(ctx, k)
					s.DeleteData([]string{k}, key)
				}
			}

			s.PingDBConnection(ctx)

		})
		err := os.Remove(tt.file)
		if err != nil {
			log.Print(err)
		}
	}
}
