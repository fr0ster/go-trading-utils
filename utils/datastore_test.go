package utils

import (
	"encoding/json"
	"os"
	"testing"

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
	ds := &DataStore{FilePath: tmpfile.Name()}

	// Create a sample data record
	data := &DataRecord{
		Balance:       100.0,
		MiddlePrice:   50.0,
		Quantity:      10.0,
		BoundQuantity: 5.0,
	}

	// Save the data to the data store
	err = ds.SaveData(data)
	assert.NoError(t, err)

	// Read the saved data from the file
	fileData, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)

	// Unmarshal the file data into a DataRecord struct
	var savedData DataRecord
	err = json.Unmarshal(fileData, &savedData)
	assert.NoError(t, err)

	// Assert that the saved data matches the original data
	assert.Equal(t, data, &savedData)
}

func TestLoadData(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "datastore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Initialize the DataStore with the temporary file path
	ds := &DataStore{FilePath: tmpfile.Name()}

	// Create a sample data record
	data := &DataRecord{
		Balance:       100.0,
		MiddlePrice:   50.0,
		Quantity:      10.0,
		BoundQuantity: 5.0,
	}

	// Marshal the data record into JSON
	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)

	// Write the JSON data to the file
	err = os.WriteFile(tmpfile.Name(), jsonData, 0644)
	assert.NoError(t, err)

	// Load the data from the data store
	loadedData, err := ds.LoadData()
	assert.NoError(t, err)

	// Assert that the loaded data matches the original data
	assert.Equal(t, data, loadedData)
}
