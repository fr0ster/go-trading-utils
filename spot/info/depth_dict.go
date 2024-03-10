package info

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/utils"
)

type DepthMapType map[Price]DepthRecord

var (
	depthMap = make(DepthMapType)
	mu_dict  sync.Mutex
)

func DepthDictMutexLock() {
	mu_dict.Lock()
}

func DepthDictMutexUnlock() {
	mu_dict.Unlock()
}

func InitDepthMap(client *binance.Client, symbolname string) (err error) {
	res, err :=
		client.NewDepthService().
			Symbol(string(symbolname)).
			Do(context.Background())
	if err != nil {
		return
	}
	mu_dict.Lock()
	defer mu_dict.Unlock()
	for _, bid := range res.Bids {
		value, exists := depthMap[Price(utils.ConvStrToFloat64(bid.Price))]
		if exists {
			value.BidQuantity += Price(utils.ConvStrToFloat64(bid.Quantity))
		} else {
			depthMap[Price(utils.ConvStrToFloat64(bid.Price))] =
				DepthRecord{
					Price(utils.ConvStrToFloat64(bid.Price)),
					res.LastUpdateID,
					0,
					res.LastUpdateID,
					Price(utils.ConvStrToFloat64(bid.Quantity)),
				}
		}
	}
	for _, ask := range res.Asks {
		value, exists := depthMap[Price(utils.ConvStrToFloat64(ask.Price))]
		if exists {
			value.AskQuantity += Price(utils.ConvStrToFloat64(ask.Quantity))
		} else {
			depthMap[Price(utils.ConvStrToFloat64(ask.Price))] =
				DepthRecord{
					Price(utils.ConvStrToFloat64(ask.Price)),
					res.LastUpdateID,
					Price(utils.ConvStrToFloat64(ask.Quantity)),
					res.LastUpdateID,
					0,
				}
		}
	}
	return nil
}

func GetDepthMap() *DepthMapType {
	mu_dict.Lock()
	defer mu_dict.Unlock()
	return &depthMap
}

func SetDepthMap(dict *DepthMapType) {
	mu_dict.Lock()
	defer mu_dict.Unlock()
	depthMap = *dict
}

func SearchDepthMap(key Price) (DepthRecord, bool) {
	mu_dict.Lock()
	defer mu_dict.Unlock()
	value, exists := depthMap[key]
	return value, exists
}

func SearchDepthMapByPrices(low Price, high Price) DepthMapType {
	mu_dict.Lock()
	defer mu_dict.Unlock()
	result := make(DepthMapType)
	for k, v := range depthMap {
		if k >= low && k <= high {
			result[k] = v
		}
	}
	return result
}

func ShowDepthMap() {
	mu_dict.Lock()
	defer mu_dict.Unlock()
	// Створюємо зріз ключів
	keys := make([]Price, 0, len(depthMap))
	for k := range depthMap {
		keys = append(keys, k)
	}

	// Сортуємо зріз ключів
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	// Проходимося по відсортованому зрізу ключів і отримуємо відповідні значення з мапи
	fmt.Println("BookTickerMap:", "Time:", time.Now().Format("2006-01-02 15:04:05"))
	for _, key := range keys {
		value := depthMap[key]
		fmt.Println(
			"Price:", key,
			"AskLastUpdateID:", value.AskLastUpdateID,
			"AskQuantity:", value.AskQuantity,
			"BidLastUpdateID:", value.BidLastUpdateID,
			"BidQuantity:", value.BidQuantity)
	}
}
