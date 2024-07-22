package kline

import (
	klines_types "github.com/fr0ster/go-trading-utils/types/klines"
	"github.com/google/btree"
)

type (
	Klines interface {
		Lock()
		Unlock()
		Ascend(func(btree.Item) bool)
		Descend(func(btree.Item) bool)
		SetKline(value *klines_types.Kline)
		GetKlines() *btree.BTree
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
)
