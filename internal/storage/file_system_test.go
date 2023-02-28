package storage

import (
	"context"
	"log"
	"os"
	"reflect"
	"testing"
)

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
			s.SaveData(context.Background(), tt.want)
			if got := s.ReadData(context.Background()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadData() = %v, want %v", got, tt.want)
			}
		})
		err := os.Remove(tt.file)
		if err != nil {
			log.Print(err)
		}
	}
}
