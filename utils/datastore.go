package utils

import (
	"os"
	"sync"
)

type (
	DataStore struct {
		FilePath string
		Data     []byte
		Mutex    sync.Mutex
	}
)

// NewDataStore is a constructor for DataStore
func NewDataStore(filePath string) *DataStore {
	return &DataStore{
		FilePath: filePath,
		Data:     []byte{},
		Mutex:    sync.Mutex{},
	}
}

func (ds *DataStore) GetData() (data []byte) {
	return ds.Data
}

func (ds *DataStore) SetData(data []byte) {
	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()

	ds.Data = data
}

// SaveToFile saves the DataStore to a file
func (ds *DataStore) SaveToFile() error {
	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()

	return os.WriteFile(ds.FilePath, ds.Data, 0644)
}

// LoadFromFile loads the DataStore from a file
func (ds *DataStore) LoadFromFile() error {
	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()

	data, err := os.ReadFile(ds.FilePath)
	if err != nil {
		return err
	}

	ds.Data = data
	return nil
}
