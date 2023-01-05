package storage

type MemoryWork struct {
	UserData map[string]map[string]string
}

func (s MemoryWork) ReadData() map[string]map[string]string {

	return s.UserData
}

func (s MemoryWork) SaveData(d map[string]map[string]string) {

	s.UserData = d

}