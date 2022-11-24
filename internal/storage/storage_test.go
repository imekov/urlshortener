package storage

import (
	"log"
	"os"
	"reflect"
	"testing"
)

func TestStorage_WriteReadData(t *testing.T) {
	tests := []struct {
		name string
		file string
		want map[string]string
	}{
		{"string numbers", "tg3948he.gob", map[string]string{"12345": "6543"}},
		{"long string", "1Jo$@gid%fg.gob", map[string]string{"hgutrhgitrhgoiwejoirjwoeijgeiojgoierg": "oigrjtohijroithjoirtjhoirtjhoirtjhoirjtioh"}},
		{"mix", "32_489.gob", map[string]string{"8394ht98ghrfjuidrjf8943u": "65gi43hhfr&^#Grh2"}},
		{"cyrillic", "hj4589gerio.gob", map[string]string{"проверка": "связи"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Storage{
				Filename: tt.file,
			}
			s.SaveData(tt.want)
			if got := s.ReadData(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadData() = %v, want %v", got, tt.want)
			}
		})
		err := os.Remove(tt.file)
		if err != nil {
			log.Fatal(err)
		}
	}
}
