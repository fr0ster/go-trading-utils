package utils

import (
	"encoding/json"
	"os"

	"github.com/adshao/go-binance/v2"
)

type DataRecord struct {
	AccountType   binance.AccountType
	Symbol        binance.SymbolType
	Balance       float64
	MiddlePrice   float64
	Quantity      float64
	BoundQuantity float64
}

// DataStore represents the data store for your program
type DataStore struct {
	FilePath string
}

// SaveData saves the data to the data store file in JSON format
func (ds *DataStore) SaveData(records []DataRecord) error {
	file, err := os.Create(ds.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(records); err != nil {
		return err
	}

	return nil
}

// LoadData loads the data from the data store file in JSON format
func (ds *DataStore) LoadData() ([]DataRecord, error) {
	file, err := os.Open(ds.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []DataRecord
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&records); err != nil {
		return nil, err
	}

	return records, nil
}
