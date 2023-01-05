package storage

type MemoryWork struct {
	State map[string]map[string]string
}

func (s MemoryWork) ReadData() map[string]map[string]string {

	return s.State
}

func (s MemoryWork) SaveData(d map[string]map[string]string) {

	s.State = d

}
