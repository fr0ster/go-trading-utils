package bookticker

import (
	"context"

	"github.com/adshao/go-binance/v2"
	bookticker_types "github.com/fr0ster/go-trading-utils/types/bookticker"
)

func Init(btt *bookticker_types.BookTickerBTree, pair string, client *binance.Client) (err error) {
	btt.Lock()         // Locking the bookticker
	defer btt.Unlock() // Unlocking the bookticker
	bookTickerList, err :=
		client.NewListBookTickersService().
			Symbol(pair).
			Do(context.Background())
	if err != nil {
		return
	}
	for _, bookTicker := range bookTickerList {
		bookTicker, err := bookticker_types.Binance2BookTicker(bookTicker)
		if err != nil {
			return err
		}
		btt.Set(bookTicker)
	}
	return nil
}
