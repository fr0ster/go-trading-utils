package kline

import (
	"github.com/google/btree"
)

type (
	Klines interface {
		Lock()
		Unlock()
		Init(apt_key, secret_key, symbolname string, UseTestnet bool)
		GetItem(openTime int64) *Kline
		SetItem(value Kline)
		Show()
	}
	// WsKline define websocket kline
	WsKline struct {
		StartTime            int64  `json:"t"`
		EndTime              int64  `json:"T"`
		Symbol               string `json:"s"`
		Interval             string `json:"i"`
		FirstTradeID         int64  `json:"f"`
		LastTradeID          int64  `json:"L"`
		Open                 string `json:"o"`
		Close                string `json:"c"`
		High                 string `json:"h"`
		Low                  string `json:"l"`
		Volume               string `json:"v"`
		TradeNum             int64  `json:"n"`
		IsFinal              bool   `json:"x"`
		QuoteVolume          string `json:"q"`
		ActiveBuyVolume      string `json:"V"`
		ActiveBuyQuoteVolume string `json:"Q"`
	}
	Kline struct {
		OpenTime                 int64  `json:"openTime"`
		Open                     string `json:"open"`
		High                     string `json:"high"`
		Low                      string `json:"low"`
		Close                    string `json:"close"`
		Volume                   string `json:"volume"`
		CloseTime                int64  `json:"closeTime"`
		QuoteAssetVolume         string `json:"quoteAssetVolume"`
		TradeNum                 int64  `json:"tradeNum"`
		TakerBuyBaseAssetVolume  string `json:"takerBuyBaseAssetVolume"`
		TakerBuyQuoteAssetVolume string `json:"takerBuyQuoteAssetVolume"`
	}
)

// Kline - тип для зберігання свічок
func (i *Kline) Less(than btree.Item) bool {
	return i.OpenTime < than.(*Kline).OpenTime
}

func (i *Kline) Equal(than btree.Item) bool {
	return i.OpenTime == than.(*Kline).OpenTime
}
