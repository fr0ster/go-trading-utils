package utils

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/stretchr/testify/assert"
)

func TestSaveData(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "datastore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Initialize the DataStore with the temporary file path
	ds := DataStore{FilePath: tmpfile.Name()}

	// Create a sample data record
	data := DataItem{
		AccountType:   binance.AccountTypeSpot,
		Symbol:        binance.SymbolType("BTCUSDT"),
		Balance:       1000.0,
		Value:         50000.0,
		Quantity:      0.02,
		BoundQuantity: 0.01,
	}

	// Save the data to the data store
	err = ds.SaveData(data)
	assert.NoError(t, err)

	// Read the saved data from the file
	fileData, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)

	// Unmarshal the JSON data into a DataRecord struct
	var savedData DataItem
	err = json.Unmarshal(fileData, &savedData)
	assert.NoError(t, err)

	// Assert that the saved data matches the original data
	assert.Equal(t, data, savedData)
}

func TestLoadData(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "datastore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Initialize the DataStore with the temporary file path
	ds := DataStore{FilePath: tmpfile.Name()}

	// Create a sample data record
	data := DataItem{
		AccountType:   binance.AccountTypeSpot,
		Symbol:        binance.SymbolType("BTCUSDT"),
		Balance:       1000.0,
		Value:         50000.0,
		Quantity:      0.02,
		BoundQuantity: 0.01,
	}

	// Save the data to the data store
	err = ds.SaveData(data)
	assert.NoError(t, err)

	// Load the data from the data store
	loadedData, err := ds.LoadData()
	assert.NoError(t, err)

	// Assert that the loaded data matches the original data
	assert.Equal(t, data, loadedData)
}
