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
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Initialize the DataStore with the temporary file path
	ds := utils.NewDataStore(tmpfile.Name())

	// Create a sample data record
	ds.AddItem(utils.DataItem("Test data string 001"))
	ds.AddItem(utils.DataItem("Test data string 002"))
	ds.AddItem(utils.DataItem("Test data string 003"))

	// Save the data to the data store
	err = ds.SaveToFile()
	assert.NoError(t, err)

	// Read the saved data from the file
	fileData, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)

	// Unmarshal the JSON data into a DataRecord struct
	var savedData []utils.DataItem
	err = json.Unmarshal(fileData, &savedData)
	assert.NoError(t, err)

	for raw, test := range savedData {
		// Assert that the saved data matches the original data
		if raw == 0 {
			assert.Equal(t, "Test data string 001", string(test))
		} else if raw == 1 {
			assert.Equal(t, "Test data string 002", string(test))
		} else if raw == 2 {
			assert.Equal(t, "Test data string 003", string(test))
		}
	}
}

func TestLoadData(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "datastore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Initialize the DataStore with the temporary file path
	ds := utils.NewDataStore(tmpfile.Name())

	// Save the data to the data store
	testData := []utils.DataItem{
		"Test data string 001",
		"Test data string 002",
		"Test data string 003",
	}

	// Assign the result of append to a variable
	json4save, err := json.Marshal(testData)

	os.WriteFile(ds.FilePath, json4save, 0644)
	assert.NoError(t, err)

	// Load the data from the file
	err = ds.LoadFromFile()
	assert.NoError(t, err)

	for raw, test := range ds.Data {
		// Assert that the loaded data matches the original data
		// Assert that the saved data matches the original data
		if raw == 0 {
			assert.Equal(t, "Test data string 001", string(test))
		} else if raw == 1 {
			assert.Equal(t, "Test data string 002", string(test))
		} else if raw == 2 {
			assert.Equal(t, "Test data string 003", string(test))
		}
	}
}
