package utils_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/fr0ster/go-binance-utils/utils"
	"github.com/stretchr/testify/assert"
)

func TestSaveData(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "datastore_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Initialize the DataStore with the temporary file path
	ds := &utils.DataStore{
		FilePath: tmpfile.Name(),
	}

	// Create some sample data
	data := []*utils.DataRecord{
		{
			Balance:       100.0,
			Orders:        nil,
			MiddlePrice:   50.0,
			Quantity:      10.0,
			BoundQuantity: 5.0,
		},
		{
			Balance:       200.0,
			Orders:        nil,
			MiddlePrice:   75.0,
			Quantity:      20.0,
			BoundQuantity: 10.0,
		},
	}

	// Save the data
	err = ds.SaveData(data)
	assert.NoError(t, err)

	// Read the saved data from the file
	fileData, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)

	// Unmarshal the JSON data
	var savedData []*utils.DataRecord
	err = json.Unmarshal(fileData, &savedData)
	assert.NoError(t, err)

	// Assert that the saved data matches the original data
	assert.Equal(t, data, savedData)
}

func TestLoadData(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "datastore_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Write some sample data to the file
	data := []*utils.DataRecord{
		{
			Balance:       100.0,
			Orders:        nil,
			MiddlePrice:   50.0,
			Quantity:      10.0,
			BoundQuantity: 5.0,
		},
		{
			Balance:       200.0,
			Orders:        nil,
			MiddlePrice:   75.0,
			Quantity:      20.0,
			BoundQuantity: 10.0,
		},
	}
	fileData, err := json.Marshal(data)
	assert.NoError(t, err)
	err = os.WriteFile(tmpfile.Name(), fileData, 0644)
	assert.NoError(t, err)

	// Initialize the DataStore with the temporary file path
	ds := &utils.DataStore{
		FilePath: tmpfile.Name(),
	}

	// Load the data
	loadedData, err := ds.LoadData()
	assert.NoError(t, err)

	// Assert that the loaded data matches the original data
	assert.Equal(t, data, loadedData)
}
