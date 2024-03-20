package client

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"github.com/fr0ster/go-trading-utils/interfaces/client"
)

type Client struct {
	client *binance.Client
	symbol string
	limit  int
}

func New(apiKey, secretKey, symbol string, limit int, UseTestnet bool) *Client {
	binance.UseTestnet = UseTestnet
	return &Client{client: binance.NewClient(apiKey, secretKey), symbol: symbol, limit: limit}
}

// NewDepthServiceDo implements client.Client.
func (c *Client) NewDepthServiceDo() (*client.DepthResponse, error) {
	res, err := c.client.NewDepthService().Symbol(c.symbol).Limit(c.limit).Do(context.Background())
	if err != nil {
		return nil, err
	}
	return &client.DepthResponse{
		LastUpdateID: res.LastUpdateID,
		Time:         0,
		TradeTime:    0,
		Bids:         res.Bids,
		Asks:         res.Asks,
	}, nil
}

// NewGetAccountServiceDo implements client.Client.
func (c *Client) NewGetAccountServiceDo() (*client.Account, error) {
	res, err := c.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	return &client.Account{
		MakerCommission:  res.MakerCommission,
		TakerCommission:  res.TakerCommission,
		BuyerCommission:  res.BuyerCommission,
		SellerCommission: res.SellerCommission,
		CanTrade:         res.CanTrade,
		CanWithdraw:      res.CanWithdraw,
		CanDeposit:       res.CanDeposit,
		UpdateTime:       res.UpdateTime,
		AccountType:      res.AccountType,
		Balances:         convertBalances(res.Balances),
		Permissions:      res.Permissions,
	}, nil
}

func convertBalances(balances []binance.Balance) []client.Balance {
	convBalances := make([]client.Balance, len(balances))
	for i, balance := range balances {
		convBalances[i] = client.Balance{
			Asset:  balance.Asset,
			Free:   balance.Free,
			Locked: balance.Locked,
		}
	}
	return convBalances
}

// NewPingServiceDo implements client.Client.
func (c *Client) NewPingServiceDo() (err error) {
	return c.client.NewPingService().Do(context.Background())
}

// NewServerTimeServiceDo implements client.Client.
func (c *Client) NewServerTimeServiceDo() {
	panic("unimplemented")
}
