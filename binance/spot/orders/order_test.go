package orders_test

import (
	"context"
	"log"
	"math"
	"os"
	"testing"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
	exchange_info "github.com/fr0ster/go-trading-utils/binance/spot/exchangeinfo"
	"github.com/fr0ster/go-trading-utils/binance/spot/orders"
	exchange_types "github.com/fr0ster/go-trading-utils/types/exchangeinfo"
	symbol_info "github.com/fr0ster/go-trading-utils/types/symbol"
	"github.com/fr0ster/go-trading-utils/utils"
)

const (
	errorMsg = "Error: %v"
	pair     = "SUSHIUSDT"
)

func GetPrice(client *binance.Client, symbol string) (float64, error) {
	price, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		return 0, err
	}
	return utils.ConvStrToFloat64(price[0].Price), nil
}

func TestNewLimitOrder(t *testing.T) {
	api_key := os.Getenv("SPOT_TEST_BINANCE_API_KEY")
	secret_key := os.Getenv("SPOT_TEST_BINANCE_SECRET_KEY")
	binance.UseTestnet = true
	client := binance.NewClient(api_key, secret_key)

	exchangeInfo := exchange_types.New()
	err := exchange_info.Init(exchangeInfo, 3, client)
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	pairInfo, err := exchangeInfo.GetSymbol(&symbol_info.SpotSymbol{Symbol: pair}).(*symbol_info.SpotSymbol).GetSpotSymbol()
	if err != nil {
		log.Printf(errorMsg, err)
		return
	}
	minQuantity := utils.ConvStrToFloat64(pairInfo.NotionalFilter().MinNotional)
	quantityRound := int(math.Log10(1 / utils.ConvStrToFloat64(pairInfo.LotSizeFilter().StepSize)))
	priceRound := int(math.Log10(1 / utils.ConvStrToFloat64(pairInfo.PriceFilter().TickSize)))
	price, err := GetPrice(client, pair)
	if err != nil {
		log.Fatalf("Error getting price: %v", err)
	}
	minQuantityStr := utils.ConvFloat64ToStr(minQuantity+1, quantityRound)
	priceStr := utils.ConvFloat64ToStr(price*0.95, priceRound)
	// Create a new limit order
	order, err := orders.NewLimitOrder(client, pair, binance.SideTypeBuy, minQuantityStr, priceStr, binance.TimeInForceTypeGTC)
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
	if order.Side != binance.SideTypeBuy {
		t.Errorf("Expected side to be Buy, got %s", order.Side)
	}
	if utils.ConvStrToFloat64(order.ExecutedQuantity) != utils.ConvStrToFloat64(minQuantityStr) && order.Status == binance.OrderStatusTypeFilled {
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
