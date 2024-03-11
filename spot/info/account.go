package info

import (
	"context"

	"github.com/adshao/go-binance/v2"
)

func GetAccountInfo(client *binance.Client) (res *binance.Account, err error) {
	res, err = client.NewGetAccountService().Do(context.Background())
	return
}

func GetBalances(client *binance.Client) (res []binance.Balance, err error) {
	accountInfo, err := GetAccountInfo(client)
	if err != nil {
		return
	}
	res = accountInfo.Balances
	return
}
