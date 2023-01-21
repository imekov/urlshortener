package storage

type MemoryWork struct {
	UserData map[string]map[string]string
}

func (s MemoryWork) ReadData() map[string]map[string]string {
	return s.UserData
}

func (s MemoryWork) SaveData(d map[string]map[string]string) error {

	for userID, values := range d {
		if _, ok := s.UserData[userID]; !ok {
			s.UserData[userID] = map[string]string{}
		}

		for shortURL, originalURL := range values {
			s.UserData[userID][shortURL] = originalURL
		}

	}

	return nil

}

func (s MemoryWork) DeleteData([]string, string) {
}

func (s MemoryWork) GetURLByShortname(shortname string) (string, bool) {

	for _, value := range s.UserData {
		if originalURL, ok := value[shortname]; ok {
			return originalURL, false
		}
	}

	return "", false
}
