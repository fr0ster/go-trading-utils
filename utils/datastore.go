package utils

import (
	"encoding/json"
	"os"

	"github.com/adshao/go-binance/v2"
)

type DataRecord struct {
	Balance       float64
	Orders        []*binance.CreateOrderResponse
	MiddlePrice   float64
	Quantity      float64
	BoundQuantity float64
}

// DataStore represents the data store for your program
type DataStore struct {
	FilePath string
}

// SaveData saves the data to the specified file path in JSON format
func (ds *DataStore) SaveData(data []*DataRecord) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(ds.FilePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// LoadData loads the data from the specified file path in JSON format
func (ds *DataStore) LoadData() ([]*DataRecord, error) {
	jsonData, err := os.ReadFile(ds.FilePath)
	if err != nil {
		return nil, err
	}

	var data []*DataRecord
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
