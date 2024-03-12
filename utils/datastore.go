package utils

import (
	"encoding/gob"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/google/btree"
)

type DataItem struct {
	Timestamp         time.Time
	AccountType       binance.AccountType
	Symbol            binance.SymbolType
	Balance           float64
	CalculatedBalance float64
	Quantity          float64
	Value             float64
	BoundQuantity     float64
}

// DataStore represents the data store for your program
type (
	DataStore struct {
		FilePath string
	}
)

var (
	dataTree *btree.BTree
	mu_file  sync.Mutex
)

// Less defines the comparison method for BookTickerItem.
// It compares the symbols of two BookTickerItems.
func (b DataItem) Less(than btree.Item) bool {
	return b.Symbol < than.(DataItem).Symbol
}

// SaveData saves the data to the data store file in JSON format
func (ds *DataStore) SaveData(data DataItem) error {
	file, err := os.Create(ds.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data.Timestamp = time.Now()
	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	return nil
}

// LoadData loads the data from the data store file
func (ds *DataStore) LoadData() (DataItem, error) {
	var data DataItem

	file, err := os.Open(ds.FilePath)
	if err != nil {
		return data, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func AddRecordToTree(record DataItem) *btree.BTree {
	mu_file.Lock()
	defer mu_file.Unlock()
	if dataTree == nil {
		dataTree = btree.New(2)
	}
	dataTree.ReplaceOrInsert(record)
	return dataTree
}

func RemoveRecordFromTree(record DataItem) *btree.BTree {
	mu_file.Lock()
	defer mu_file.Unlock()
	if dataTree == nil {
		return nil
	}
	dataTree.Delete(record)
	return dataTree
}

func GetRecordFromTree(symbol binance.SymbolType) *DataItem {
	mu_file.Lock()
	defer mu_file.Unlock()
	if dataTree == nil {
		return nil
	}
	item := dataTree.Get(DataItem{Symbol: symbol})
	if item == nil {
		return nil
	}
	return item.(*DataItem)
}

func GetTree() *btree.BTree {
	mu_file.Lock()
	defer mu_file.Unlock()
	return dataTree
}

func SetTree(tree *btree.BTree) {
	mu_file.Lock()
	defer mu_file.Unlock()
	dataTree = tree
}

func (ds *DataStore) SaveTreeToFile(filePath string) error {
	mu_file.Lock()
	defer mu_file.Unlock()
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(dataTree)
	if err != nil {
		return err
	}
	return nil
}

func (ds *DataStore) LoadTreeFromFile(filePath string) error {
	mu_file.Lock()
	defer mu_file.Unlock()
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(dataTree)
	if err != nil {
		return err
	}
	return nil
}
