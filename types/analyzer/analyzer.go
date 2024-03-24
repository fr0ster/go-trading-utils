package main

import (
	"context"

	"github.com/adshao/go-binance/v2"
	config_interfaces "github.com/fr0ster/go-trading-utils/interfaces/config"
	types "github.com/fr0ster/go-trading-utils/types"
)

type (
	Analyzer struct {
		SpotClient     *binance.Client
		SpotConfig     config_interfaces.Configuration
		TargetPrice    float64
		TargetQuantity float64
		Levels         []types.DepthLevels
		limits         int
	}
)

// GetSide implements Analyzers.
func (a *Analyzer) GetSide() (types.OrderSide, error) {
	panic("unimplemented")
}

func (a *Analyzer) GetLevels() ([]types.DepthLevels, error) {
	depth, err := a.SpotClient.NewDepthService().Symbol(a.SpotConfig.GetSymbol()).Limit(a.limits).Do(context.Background())
	if err != nil {
		return nil, err
	}
	for _, bid := range depth.Bids {
		price, quantity, _ := bid.Parse()
		a.Levels = append(a.Levels, types.DepthLevels{
			Price:    price,
			Side:     types.DepthSideBid,
			Quantity: quantity,
		})
	}
	for _, ask := range depth.Asks {
		price, quantity, _ := ask.Parse()
		a.Levels = append(a.Levels, types.DepthLevels{
			Price:    price,
			Side:     types.DepthSideAsk,
			Quantity: quantity,
		})
	}
	return a.Levels, nil
}

func NewAnalyzer(spotClient *binance.Client, spotConfig config_interfaces.Configuration, limits int) *Analyzer {
	return &Analyzer{
		SpotClient: spotClient,
		SpotConfig: spotConfig,
		Levels:     make([]types.DepthLevels, 0),
		limits:     limits,
	}
}
