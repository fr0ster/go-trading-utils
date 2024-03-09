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

// SaveData saves the data to the data store file in JSON format
func (ds *DataStore) SaveData(data *DataRecord) error {
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

// LoadData loads the data from the data store file in JSON format
func (ds *DataStore) LoadData() (*DataRecord, error) {
	file, err := os.Open(ds.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	data := &DataRecord{}
	err = decoder.Decode(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
