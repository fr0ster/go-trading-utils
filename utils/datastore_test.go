package utils

import (
	"encoding/json"
	"io/ioutil"

	// "io/ioutil"
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
	ds := DataStore{FilePath: tmpfile.Name()}

	// Create some sample data records
	records := []DataRecord{
		{AccountType: "spot", Symbol: "BTCUSDT", Balance: 1.0, MiddlePrice: 50000.0, Quantity: 0.5, BoundQuantity: 0.2},
		{AccountType: "margin", Symbol: "ETHUSDT", Balance: 2.0, MiddlePrice: 3000.0, Quantity: 1.0, BoundQuantity: 0.5},
	}

	// Save the data records
	err = ds.SaveData(records)
	assert.NoError(t, err)

	// Read the saved data from the file
	data, err := os.ReadFile(tmpfile.Name())
	assert.NoError(t, err)

	// Unmarshal the JSON data into a slice of DataRecord
	var savedRecords []DataRecord
	err = json.Unmarshal(data, &savedRecords)
	assert.NoError(t, err)

	// Assert that the saved records match the original records
	assert.Equal(t, records, savedRecords)
}

func TestLoadData(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := ioutil.TempFile("", "datastore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Initialize the DataStore with the temporary file path
	ds := DataStore{FilePath: tmpfile.Name()}

	// Create some sample data records
	records := []DataRecord{
		{AccountType: "spot", Symbol: "BTCUSDT", Balance: 1.0, MiddlePrice: 50000.0, Quantity: 0.5, BoundQuantity: 0.2},
		{AccountType: "margin", Symbol: "ETHUSDT", Balance: 2.0, MiddlePrice: 3000.0, Quantity: 1.0, BoundQuantity: 0.5},
	}

	// Marshal the data records into JSON
	data, err := json.Marshal(records)
	assert.NoError(t, err)

	// Write the JSON data to the file
	err = ioutil.WriteFile(tmpfile.Name(), data, 0644)
	assert.NoError(t, err)

	// Load the data records
	loadedRecords, err := ds.LoadData()
	assert.NoError(t, err)

	// Assert that the loaded records match the original records
	assert.Equal(t, records, loadedRecords)
}
