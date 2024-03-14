package utils

import (
	"encoding/json"
	"os"
	"sync"
)

type (
	DataItem  string
	DataStore struct {
		FilePath string
		Data     []DataItem
		Mutex    sync.Mutex
	}
)

// NewDataStore is a constructor for DataStore
func NewDataStore(filePath string) *DataStore {
	return &DataStore{
		FilePath: filePath,
		Data:     []DataItem{},
		Mutex:    sync.Mutex{},
	}
}

// AddItem adds a new DataItem to the DataStore
func (ds *DataStore) AddItem(item DataItem) {
	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()

	ds.Data = append(ds.Data, []DataItem{item}...)
}

// SaveToFile saves the DataStore to a file
func (ds *DataStore) SaveToFile() error {
	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()

	data, err := json.Marshal(ds.Data)
	if err != nil {
		return err
	}

	return os.WriteFile(ds.FilePath, data, 0644)
}

// LoadFromFile loads the DataStore from a file
func (ds *DataStore) LoadFromFile() error {
	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()

	data, err := os.ReadFile(ds.FilePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &ds.Data)
}
