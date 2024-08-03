package processor

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/fr0ster/go-trading-utils/types"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/sirupsen/logrus"

	futures_exchangeinfo "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"

	items_types "github.com/fr0ster/go-trading-utils/types/depths/items"
	exchangeinfo_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	processor_types "github.com/fr0ster/go-trading-utils/types/processor"
)

func New(
	client *futures.Client,
	degree int,
	symbol string,
	limitOnPosition items_types.ValueType,
	limitOnTransaction items_types.ValuePercentType,
	UpAndLowBound items_types.PricePercentType,
	deltaPrice items_types.PricePercentType,
	deltaQuantity items_types.QuantityPercentType,
	leverage int,
	marginType types.MarginType,
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
	exchange := exchangeinfo_types.New(futures_exchangeinfo.InitCreator(client, degree, symbol))
	symbolInfo := exchange.GetSymbol(symbol)
	baseSymbol := string(exchange.GetSymbol(symbol).GetBaseSymbol())
	targetSymbol := string(exchange.GetSymbol(symbol).GetTargetSymbol())
	pairProcessor, err = processor_types.New(
		quit,       // quit
		symbol,     // pair
		symbolInfo, // symbolInfo
		getBaseBalance(
			client,     // client
			baseSymbol, // symbol
		), // getBaseBalance
		getTargetBalance(
			client,       // client
			targetSymbol, // symbol
		), // getTargetBalance
		getFreeBalance(
			client,     // client
			baseSymbol, // symbol
		), // getFreeBalance
		getLockedBalance(
			client,     // client
			baseSymbol, // symbol
		), // getLockedBalance
		getCurrentPrice(client, symbol), // getCurrentPrice
		getPositionRisk(client),         // getPositionRisk
		func() int {
			return leverage
		}, // getLeverage
		setLeverage(client), // setLeverage
		func() types.MarginType {
			return marginType
		}, // getMarginType
		setMarginType(client),     // setMarginType
		setPositionMargin(client), // setPositionMargin
		// closePosition(),           // closePosition
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

func getBaseBalance(client *futures.Client, symbol string) processor_types.GetBaseBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Assets {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance))
			}
		}
		return 0.0
	}
} // getBaseBalance
func getTargetBalance(client *futures.Client, symbol string) processor_types.GetTargetBalanceFunction {
	return func() items_types.QuantityType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Assets {
			if asset.Asset == symbol {
				return items_types.QuantityType(utils.ConvStrToFloat64(asset.WalletBalance))
			}
		}
		return 0.0
	}
} // getTargetBalance
func getFreeBalance(client *futures.Client, symbol string) processor_types.GetFreeBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Assets {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.AvailableBalance))
			}
		}
		return 0.0
	}
} // getFreeBalance
func getLockedBalance(client *futures.Client, symbol string) processor_types.GetLockedBalanceFunction {
	return func() items_types.ValueType {
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get account: %v", err)
			return 0
		}
		for _, asset := range account.Assets {
			if asset.Asset == symbol {
				return items_types.ValueType(utils.ConvStrToFloat64(asset.WalletBalance) - utils.ConvStrToFloat64(asset.AvailableBalance))
			}
		}
		return 0.0
	}
} // getLockedBalance
func getCurrentPrice(client *futures.Client, symbol string) processor_types.GetCurrentPriceFunction {
	return func() items_types.PriceType {
		price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
		if err != nil {
			logrus.Errorf("Can't get price: %v", err)
			return 0
		}
		return items_types.PriceType(utils.ConvStrToFloat64(price[0].Price))
	}
} // getCurrentPrice
func getPositionRisk(client *futures.Client) func(*processor_types.Processor) processor_types.GetPositionRiskFunction {
	return func(p *processor_types.Processor) processor_types.GetPositionRiskFunction {
		return func() *futures.PositionRisk {
			risks, err := client.NewGetPositionRiskService().Symbol(p.GetSymbol()).Do(context.Background())
			if err == nil {
				return risks[0]
			}
			return &futures.PositionRisk{}
		}
	}
} // getPositionRisk
func setLeverage(client *futures.Client) func(p *processor_types.Processor) processor_types.SetLeverageFunction {
	return func(p *processor_types.Processor) processor_types.SetLeverageFunction {
		return func(leverage int) (Leverage int, MaxNotionalValue string, Symbol string, err error) {
			var res *futures.SymbolLeverage
			res, err = client.NewChangeLeverageService().Symbol(p.GetSymbol()).Leverage(leverage).Do(context.Background())
			Leverage = res.Leverage
			MaxNotionalValue = res.MaxNotionalValue
			Symbol = res.Symbol
			return
		}
	}
} // setLeverage
func setMarginType(client *futures.Client) func(p *processor_types.Processor) processor_types.SetMarginTypeFunction {
	return func(p *processor_types.Processor) processor_types.SetMarginTypeFunction {
		return func(marginType types.MarginType) error {
			return client.NewChangeMarginTypeService().Symbol(p.GetSymbol()).MarginType(futures.MarginType(marginType)).Do(context.Background())
		}
	}
} // setMarginType
func setPositionMargin(client *futures.Client) func(p *processor_types.Processor) processor_types.SetPositionMarginFunction {
	return func(p *processor_types.Processor) processor_types.SetPositionMarginFunction {
		return func(amountMargin items_types.ValueType, typeMargin int) error {
			return client.
				NewUpdatePositionMarginService().
				Symbol(p.GetSymbol()).
				Amount(utils.ConvFloat64ToStrDefault(float64(amountMargin))).
				Type(typeMargin).
				Do(context.Background())
		}
	}
} // setPositionMargin
