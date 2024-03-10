package info

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/utils"
)

type DepthMapType map[Price]DepthRecord

var (
	depthMap = make(DepthMapType)
	mu_dict  *sync.Mutex
)

func InitDepthMap(client *binance.Client, mu *sync.Mutex, symbolname string) (err error) {
	mu_dict = mu
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
