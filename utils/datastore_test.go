package utils_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndLoadData(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "datastore_test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Initialize the DataStore with the temporary file path
	ds := utils.NewDataStore(tmpfile.Name())

	// Create a sample data record
	data := utils.DataItem{
		Timestamp:     time.Now().Unix(),
		AccountType:   binance.AccountTypeSpot,
		Symbol:        binance.SymbolType("BTCUSDT"),
		Balance:       1000.0,
		Value:         50000.0,
		Quantity:      0.02,
		BoundQuantity: 0.01,
	}
	ds.AddData(data)

	// Save the data to the data store
	err = ds.SaveData()
	assert.NoError(t, err)

	// Read the saved data from the file
	fileData, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)

	// Unmarshal the JSON data into a DataRecord struct
	var savedData utils.DataItem
	for _, item := range fileData {
		dataBytes := []byte(string(item))
		err = json.Unmarshal(dataBytes, &savedData)
		assert.NoError(t, err)

		// Assert that the saved data matches the original data

		assert.Equal(t, data.AccountType, savedData.AccountType)
		assert.Equal(t, data.Symbol, savedData.Symbol)
		assert.Equal(t, data.Balance, savedData.Balance)
		assert.Equal(t, data.Value, savedData.Value)
		assert.Equal(t, data.Quantity, savedData.Quantity)
		assert.Equal(t, data.BoundQuantity, savedData.BoundQuantity)
		assert.Equal(t, data.Timestamp, savedData.Timestamp)
	}
}
