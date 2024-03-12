package streams

import (
	"sync"

	"github.com/adshao/go-binance/v2"
	"github.com/google/btree"
)

const (
	ChannelUserData   = "userData"
	ChannelDepth      = "depth"
	ChannelKline      = "kline"
	ChannelTrade      = "trade"
	ChannelAggTrade   = "aggTrade"
	ChannelBookTicker = "bookTicker"
)

type EventChannelType struct {
	Name    string
	Channel interface{}
}

var (
	eventChannels   = btree.New(2) // Book ticker tree
	mu_eventChannel sync.Mutex     // Mutex for book ticker tree
)

// Less defines the comparison method for BookTickerItem.
// It compares the symbols of two BookTickerItems.
func (b EventChannelType) Less(than btree.Item) bool {
	return b.Name < than.(EventChannelType).Name
}

func StartUserDataStream(listenKey string, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	channel := make(chan *binance.WsUserDataEvent)
	wsHandler := func(event *binance.WsUserDataEvent) {
		channel <- event
	}
	eventChannels.ReplaceOrInsert(EventChannelType{ChannelUserData, channel})
	return binance.WsUserDataServe(listenKey, wsHandler, handleErr)
}

func GetUserDataChannel() (chan *binance.WsUserDataEvent, bool) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	item := eventChannels.Get(EventChannelType{ChannelUserData, nil})
	if item == nil {
		return nil, false
	}
	return item.(EventChannelType).Channel.(chan *binance.WsUserDataEvent), true
}

func StartDepthStream(symbol string, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	channel := make(chan *binance.WsDepthEvent)
	wsHandler := func(event *binance.WsDepthEvent) {
		channel <- event
	}
	eventChannels.ReplaceOrInsert(EventChannelType{ChannelDepth, channel})
	return binance.WsDepthServe(symbol, wsHandler, handleErr)
}

func GetDepthChannel() (chan *binance.WsDepthEvent, bool) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	item := eventChannels.Get(EventChannelType{ChannelDepth, nil})
	if item == nil {
		return nil, false
	}
	return item.(EventChannelType).Channel.(chan *binance.WsDepthEvent), true
}

func StartKlineStream(symbol string, interval string, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	channel := make(chan *binance.WsKlineEvent)
	wsHandler := func(event *binance.WsKlineEvent) {
		channel <- event
	}
	eventChannels.ReplaceOrInsert(EventChannelType{ChannelKline, channel})
	return binance.WsKlineServe(symbol, interval, wsHandler, handleErr)
}

func GetKlineChannel() (chan *binance.WsKlineEvent, bool) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	item := eventChannels.Get(EventChannelType{ChannelKline, nil})
	if item == nil {
		return nil, false
	}
	return item.(EventChannelType).Channel.(chan *binance.WsKlineEvent), true
}

func StartTradeStream(symbol string, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	channel := make(chan *binance.WsTradeEvent)
	wsHandler := func(event *binance.WsTradeEvent) {
		channel <- event
	}
	eventChannels.ReplaceOrInsert(EventChannelType{ChannelTrade, channel})
	return binance.WsTradeServe(symbol, wsHandler, handleErr)
}

func GetTradeChannel() (chan *binance.WsTradeEvent, bool) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	item := eventChannels.Get(EventChannelType{ChannelTrade, nil})
	if item == nil {
		return nil, false
	}
	return item.(EventChannelType).Channel.(chan *binance.WsTradeEvent), true
}

func StartAggTradeStream(symbol string, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	channel := make(chan *binance.WsAggTradeEvent)
	wsHandler := func(event *binance.WsAggTradeEvent) {
		channel <- event
	}
	eventChannels.ReplaceOrInsert(EventChannelType{ChannelAggTrade, channel})
	return binance.WsAggTradeServe(symbol, wsHandler, handleErr)
}

func GetAggTradeChannel() (chan *binance.WsAggTradeEvent, bool) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	item := eventChannels.Get(EventChannelType{ChannelAggTrade, nil})
	if item == nil {
		return nil, false
	}
	return item.(EventChannelType).Channel.(chan *binance.WsAggTradeEvent), true
}

func StartBookTickerStream(symbol string, handleErr binance.ErrHandler) (doneC, stopC chan struct{}, err error) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	channel := make(chan *binance.WsBookTickerEvent)
	wsHandler := func(event *binance.WsBookTickerEvent) {
		channel <- event
	}
	eventChannels.ReplaceOrInsert(EventChannelType{ChannelBookTicker, make(chan *binance.WsBookTickerEvent)})
	return binance.WsBookTickerServe(symbol, wsHandler, handleErr)
}

func GetBookTickerChannel() (chan *binance.WsBookTickerEvent, bool) {
	// mu_eventChannel.Lock()
	// defer mu_eventChannel.Unlock()
	item := eventChannels.Get(EventChannelType{ChannelBookTicker, nil})
	if item == nil {
		return nil, false
	}
	return item.(EventChannelType).Channel.(chan *binance.WsBookTickerEvent), true
}
