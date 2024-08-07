package processor

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"

	spot_exchangeinfo "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchangeinfo_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	processor_types "github.com/fr0ster/go-trading-utils/types/processor"
)

func New(
	client *binance.Client,
	degree int,
	symbol string,
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	UpAndLowBound items_types.PricePercentType,
	deltaPrice items_types.PricePercentType,
	deltaQuantity items_types.QuantityPercentType,
	callbackRate items_types.PricePercentType,
	debug bool,
	quits ...chan struct{},
) (pairProcessor *processor_types.Processor, err error) {
	var quit chan struct{}
	if len(quits) > 0 {
		quit = quits[0]
	} else {
		quit = make(chan struct{})
	}
	exchange := exchangeinfo_types.New(spot_exchangeinfo.InitCreator(client, degree, symbol))
	symbolInfo := exchange.GetSymbol(symbol)
	baseSymbol := string(exchange.GetSymbol(symbol).GetBaseSymbol())
	targetSymbol := string(exchange.GetSymbol(symbol).GetTargetSymbol())
	pairProcessor, err = processor_types.New(
		quit,       // quit
		symbol,     // pair
		symbolInfo, // symbolInfo
		getBaseBalance(
			client, // client
			baseSymbol,
		), // getBaseBalance
		getTargetBalance(
			client, // client
			targetSymbol,
		), // getTargetBalance
		getFreeBalance(
			client, // client
			symbol,
		), // getFreeBalance
		getLockedBalance(
			client, // client
			symbol,
		), // getLockedBalance
		getCurrentPrice(
			client, // client
			symbol,
		), // getCurrentPrice
		nil, // getPositionRisk
		nil, // getLeverage
		nil, // setLeverage
		nil, // getMarginType
		nil, // setMarginType
		nil, // setPositionMargin
		// nil, // closePosition
		func() items_types.PricePercentType {
			return deltaPrice
		}, // getDeltaPrice
		func() items_types.QuantityPercentType {
			return deltaQuantity
		}, // getDeltaQuantity
		func() items_types.ValueType {
			return limitOnPosition
		}, // getLimitOnPosition
		func() items_types.ValuePercentType {
			return limitOnTransaction
		}, // getLimitOnTransaction
		func() items_types.PricePercentType {
			return UpAndLowBound
		}, // getUpAndLowBound
		func() items_types.PricePercentType {
			return callbackRate
		}, // getCallbackRate
		debug)
	return
} // New

func getBaseBalance(client *binance.Client, symbol string) processor_types.GetBaseBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Balances {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.Free) + utils.ConvStrToFloat64(asset.Locked))
			}
		}
		return 0.0
	}
} // getBaseBalance
func getTargetBalance(client *binance.Client, symbol string) processor_types.GetTargetBalanceFunction {
	return func() items_types.QuantityType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Balances {
			if asset.Asset == symbol {
				return items_types.QuantityType(utils.ConvStrToFloat64(asset.Free) + utils.ConvStrToFloat64(asset.Locked))
			}
		}
		return 0.0
	}
} // getTargetBalance
func getFreeBalance(client *binance.Client, symbol string) processor_types.GetFreeBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Balances {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.Free))
			}
		}
		return 0.0
	}
} // getFreeBalance
func getLockedBalance(client *binance.Client, symbol string) processor_types.GetLockedBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Balances {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.Locked))
			}
		}
		return 0.0
	}
} // getLockedBalance
func getCurrentPrice(client *binance.Client, symbol string) processor_types.GetCurrentPriceFunction {
	return func() items_types.PriceType {
		price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get price: %v", err)
			return 0
		}
		return items_types.PriceType(utils.ConvStrToFloat64(price[0].Price))
	}
} // getCurrentPrice
