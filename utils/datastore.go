package utils

import (
	"encoding/json"
	"os"

	"github.com/adshao/go-binance/v2"
)

type DataRecord struct {
	AccountType       binance.AccountType
	Symbol            binance.SymbolType
	Balance           float64
	CalculatedBalance float64
	Quantity          float64
	Value             float64
	BoundQuantity     float64
}

// DataStore represents the data store for your program
type DataStore struct {
	FilePath string
}

// SaveData saves the data to the data store file in JSON format
func (ds *DataStore) SaveData(data DataRecord) error {
	file, err := os.Create(ds.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	return nil
}

// LoadData loads the data from the data store file
func (ds *DataStore) LoadData() (DataRecord, error) {
	var data DataRecord

	file, err := os.Open(ds.FilePath)
	if err != nil {
		return data, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}
