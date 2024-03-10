package info

import (
	"context"
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-binance-utils/spot/utils"
)

var (
	bookTickerMap = make(map[Price]BookTicker)
	mu_dict       sync.Mutex
)

func InitDepthDictMap(client *binance.Client, symbolname string) (err error) {
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
		value, exists := bookTickerMap[Price(utils.ConvStrToFloat64(bid.Price))]
		if exists {
			value.BidQuantity += Price(utils.ConvStrToFloat64(bid.Quantity))
		} else {
			bookTickerMap[Price(utils.ConvStrToFloat64(bid.Price))] =
				BookTicker{
					Price(utils.ConvStrToFloat64(bid.Price)),
					res.LastUpdateID,
					0,
					res.LastUpdateID,
					Price(utils.ConvStrToFloat64(bid.Quantity)),
				}
		}
	}
	for _, ask := range res.Asks {
		value, exists := bookTickerMap[Price(utils.ConvStrToFloat64(ask.Price))]
		if exists {
			value.AskQuantity += Price(utils.ConvStrToFloat64(ask.Quantity))
		} else {
			bookTickerMap[Price(utils.ConvStrToFloat64(ask.Price))] =
				BookTicker{
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

func GetBookTickerMap() map[Price]BookTicker {
	mu_dict.Lock()
	defer mu_dict.Unlock()
	return bookTickerMap
}

func SetBookTickerMap(dict map[Price]BookTicker) {
	mu_dict.Lock()
	defer mu_dict.Unlock()
	bookTickerMap = dict
}

func SearchBookTickerMap(key Price) (BookTicker, bool) {
	mu_dict.Lock()
	defer mu_dict.Unlock()
	value, exists := bookTickerMap[key]
	return value, exists
}

func SearchBookTickerMapByPrices(low Price, high Price) map[Price]BookTicker {
	mu_dict.Lock()
	defer mu_dict.Unlock()
	result := make(map[Price]BookTicker)
	for k, v := range bookTickerMap {
		if k >= low && k <= high {
			result[k] = v
		}
	}
	return result
}
