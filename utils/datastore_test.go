package utils_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/fr0ster/go-trading-utils/types"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/stretchr/testify/assert"
)

func TestSaveData(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "datastore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Create a sample data record
	time, _ := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	testData := types.Config{
		Timestamp:         time,
		AccountType:       "SPOT",
		Symbol:            "BTCUSDT",
		Balance:           1000.0,
		CalculatedBalance: 1000.0,
		Quantity:          1.0,
		Value:             1000.0,
		BoundQuantity:     1.0,
	}

	// Initialize the DataStore with the temporary file path
	ds := utils.NewDataStore(tmpfile.Name())

	// Save the data to the data store
	jsonData, err := json.Marshal(testData)
	assert.NoError(t, err)
	ds.SetData(jsonData)
	err = ds.SaveToFile()
	assert.NoError(t, err)

	// Read the saved data from the file
	fileData, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)

	// Unmarshal the JSON data into a DataRecord struct
	var savedData types.Config
	err = json.Unmarshal(fileData, &savedData)
	assert.NoError(t, err)

	// Assert that the saved data matches the original data
	if savedData != testData {
		t.Errorf("Expected data to be %v, but got %v", testData, savedData)
	}
}

func TestLoadData(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "datastore_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Initialize the DataStore with the temporary file path
	ds := utils.NewDataStore(tmpfile.Name())

	// Save the data to the data store
	time, _ := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	testData := types.Config{
		Timestamp:         time,
		AccountType:       "SPOT",
		Symbol:            "BTCUSDT",
		Balance:           1000.0,
		CalculatedBalance: 1000.0,
		Quantity:          1.0,
		Value:             1000.0,
		BoundQuantity:     1.0,
	}

	// Assign the result of append to a variable
	json4save, err := json.Marshal(testData)

	os.WriteFile(ds.FilePath, json4save, 0644)
	assert.NoError(t, err)

	// Load the data from the file
	err = ds.LoadFromFile()
	assert.NoError(t, err)

	var savedData types.Config
	err = json.Unmarshal(ds.GetData(), &savedData)
	assert.NoError(t, err)

	// Assert that the loaded data matches the original data
	// Assert that the saved data matches the original data
	if savedData != testData {
		t.Errorf("Expected data to be %v, but got %v", testData, savedData)
	}
}
