package kline

import (
	"sync"

	"github.com/google/btree"
)

type (
	Kline struct {
		btree.BTree
		mutex  sync.Mutex
		degree int
	}
)

// Kline - B-дерево для зберігання стакана заявок
func New(degree int) *Kline {
	return &Kline{
		BTree:  *btree.New(int(degree)),
		mutex:  sync.Mutex{},
		degree: degree,
	}
}

// Init implements depth_interface.Depths.
// func (d *Kline) Init(apt_key string, secret_key string, symbolname string, UseTestnet bool) *depth_interface.Depths {
// 	binance.UseTestnet = UseTestnet
// 	res, err :=
// 		binance.NewClient(apt_key, secret_key).NewKlinesService().
// 			Symbol(string(symbolname)).
// 			Do(context.Background())
// 	if err != nil {
// 		return nil
// 	}
// 	return nil
// }
