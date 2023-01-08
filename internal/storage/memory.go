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
