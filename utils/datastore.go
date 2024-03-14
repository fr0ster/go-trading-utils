package utils

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/adshao/go-binance/v2"
)

type (
	DataItem struct {
		Timestamp         int64
		AccountType       binance.AccountType
		Symbol            binance.SymbolType
		Balance           float64
		CalculatedBalance float64
		Quantity          float64
		Value             float64
		BoundQuantity     float64
		Msg               string
	}

	DataStore struct {
		FilePath string
		Data     []DataItem
		Mutex    sync.Mutex
	}
)

func NewDataStore(filePath string) *DataStore {
	return &DataStore{
		FilePath: filePath,
		Data:     make([]DataItem, 0),
		Mutex:    sync.Mutex{},
	}
}

func (ds *DataStore) Lock() {
	ds.Mutex.Lock()
}

func (ds *DataStore) Unlock() {
	ds.Mutex.Unlock()
}

func (ds *DataStore) AddData(data DataItem) {
	ds.Lock()
	defer ds.Unlock()
	ds.Data = append(ds.Data, data)
}

// SaveData saves the data to the data store file in JSON format
func (ds *DataStore) SaveData() error {
	file, err := os.Open(ds.FilePath)
	if err != nil {
		file, err = os.Create(ds.FilePath)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(ds.Data)
	if err != nil {
		return err
	}

	return nil
}

// LoadData loads the data from the data store file
func (ds *DataStore) LoadData() error {
	file, err := os.Open(ds.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&ds.Data)
	if err != nil {
		return err
	}

	return nil
}
