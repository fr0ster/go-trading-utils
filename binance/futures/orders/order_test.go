package orders_test

import (
	"context"
	"log"
	"math"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2/common"
	"github.com/adshao/go-binance/v2/futures"
	exchange_info "github.com/fr0ster/go-trading-utils/binance/futures/exchangeinfo"
	"github.com/fr0ster/go-trading-utils/binance/futures/orders"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	"github.com/fr0ster/go-trading-utils/utils"
)

const (
	errorMsg = "Error: %v"
	pair     = "SUSHIUSDT"
)

func GetPrice(client *futures.Client, symbol string) (float64, error) {
	price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return utils.ConvStrToFloat64(price[0].Price), nil
}

func TestNewLimitOrder(t *testing.T) {
	api_key := os.Getenv("FUTURE_TEST_BINANCE_API_KEY")
	secret_key := os.Getenv("FUTURE_TEST_BINANCE_SECRET_KEY")
	futures.UseTestnet = true
	client := futures.NewClient(api_key, secret_key)

	exchangeInfo := exchange_types.New()
	err := exchange_info.Init(exchangeInfo, 3, client)
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	pairInfo, err := exchangeInfo.GetSymbol(&symbol_info.FuturesSymbol{Symbol: pair}).(*symbol_info.FuturesSymbol).GetFuturesSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	minQuantity := utils.ConvStrToFloat64(pairInfo.MinNotionalFilter().Notional)
	quantityRound := int(math.Log10(1 / utils.ConvStrToFloat64(pairInfo.LotSizeFilter().StepSize)))
	priceRound := int(math.Log10(1 / utils.ConvStrToFloat64(pairInfo.PriceFilter().TickSize)))
	price, err := GetPrice(client, pair)
	if err != nil {
		log.Fatalf("Error getting price: %v", err)
	}
	minQuantityStr := utils.ConvFloat64ToStr(minQuantity+1, quantityRound)
	priceStr := utils.ConvFloat64ToStr(price, priceRound)
	// Create a new limit order
	order, err := orders.NewLimitOrder(client, pair, futures.SideTypeBuy, minQuantityStr, priceStr, futures.TimeInForceTypeGTC)
	if err != nil {
		if apiErr, _ := err.(*common.APIError); apiErr.Code == 0 {
			log.Printf("Error with code 0: %v", err)
			return
		} else {
			log.Fatalf("Error creating limit order: %v", err)
		}
	}

	// Verify the order details
	if order.Symbol != pair {
		t.Errorf("Expected symbol to be SUSHIUSDT, got %s", order.Symbol)
	}
	if order.Side != futures.SideTypeBuy {
		t.Errorf("Expected side to be Buy, got %s", order.Side)
	}
	if order.ExecutedQuantity != minQuantityStr && order.Status == futures.OrderStatusTypeFilled {
		t.Errorf("Expected quantity to be %s, got %s", minQuantityStr, order.ExecutedQuantity)
	}
	if utils.ConvStrToFloat64(order.Price) != utils.ConvStrToFloat64(priceStr) {
		t.Errorf("Expected price to be %s, got %s", priceStr, order.Price)
	}
	_, err = orders.CancelOrder(client, order)
	if err != nil {
		log.Fatalf("Error cancelling order: %v", err)
	}
}
