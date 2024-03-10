package info

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/utils"
	"github.com/google/btree"
)

var (
	PricesTree     = btree.New(2)
	mu_prices_tree sync.Mutex
)

func (i PriceRecord) Less(than btree.Item) bool {
	return i.Price < than.(PriceRecord).Price
}

func InitPricesTree(client *binance.Client, symbolname string) (err error) {
	res, err :=
		client.NewListPricesService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	mu_prices_dict.Lock()
	defer mu_prices_dict.Unlock()
	for _, price := range res {
		PricesTree.ReplaceOrInsert(PriceRecord{
			SymbolName: SymbolName(price.Symbol),
			Price:      SymbolPrice(utils.ConvStrToFloat64(price.Price)),
		})
	}
	return nil
}

func GetPricesTree() *btree.BTree {
	mu_prices_tree.Lock()
	defer mu_prices_tree.Unlock()
	return PricesTree
}

func SearchPricesTree(price SymbolPrice) *PriceRecord {
	mu_prices_tree.Lock()
	defer mu_prices_tree.Unlock()
	return PricesTree.Get(PriceRecord{Price: price}).(*PriceRecord)
}
