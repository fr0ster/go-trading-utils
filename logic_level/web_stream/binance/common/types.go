package common

// Визначення структури для JSON AggTrade
type AggTrade struct {
	EventType    string `json:"e"`
	EventTime    int64  `json:"E"`
	TradeID      int64  `json:"a"`
	Symbol       string `json:"s"`
	Price        string `json:"p"`
	Quantity     string `json:"q"`
	FirstTradeID int64  `json:"f"`
	LastTradeID  int64  `json:"l"`
	TradeTime    int64  `json:"T"`
	IsBuyerMaker bool   `json:"m"`
}

// Визначення структури для JSON BookTicker
type BookTicker struct {
	UpdateID     int64  `json:"u"`
	Symbol       string `json:"s"`
	BestBidPrice string `json:"b"`
	BestBidQty   string `json:"B"`
	BestAskPrice string `json:"a"`
	BestAskQty   string `json:"A"`
}

// Визначення структури для JSON Kline
type Kline struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	KlineData struct {
		StartTime       int64  `json:"t"`
		CloseTime       int64  `json:"T"`
		Symbol          string `json:"s"`
		Interval        string `json:"i"`
		FirstTradeID    int64  `json:"f"`
		LastTradeID     int64  `json:"L"`
		OpenPrice       string `json:"o"`
		ClosePrice      string `json:"c"`
		HighPrice       string `json:"h"`
		LowPrice        string `json:"l"`
		Volume          string `json:"v"`
		NumberOfTrades  int    `json:"n"`
		IsFinal         bool   `json:"x"`
		QuoteVolume     string `json:"q"`
		ActiveBuyVolume string `json:"V"`
		ActiveBuyQuote  string `json:"Q"`
		Ignore          string `json:"B"`
	} `json:"k"`
}
