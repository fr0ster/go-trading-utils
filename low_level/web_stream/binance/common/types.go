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

// Визначення структури для JSON DepthUpdate
type DepthUpdate struct {
	EventType     string     `json:"e"`
	EventTime     int64      `json:"E"`
	TransactTime  int64      `json:"T"`
	Symbol        string     `json:"s"`
	FirstUpdateID int64      `json:"U"`
	LastUpdateID  int64      `json:"u"`
	PrevUpdateID  int64      `json:"pu"`
	Bids          [][]string `json:"b"`
	Asks          [][]string `json:"a"`
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

// Визначення структури для JSON Trade
type OrderTradeUpdate struct {
	EventType    string `json:"e"`
	EventTime    int64  `json:"E"`
	TransactTime int64  `json:"T"`
	Order        struct {
		Symbol              string `json:"s"`
		ClientOrderID       string `json:"c"`
		Side                string `json:"S"`
		OrderType           string `json:"o"`
		TimeInForce         string `json:"f"`
		Quantity            string `json:"q"`
		Price               string `json:"p"`
		AveragePrice        string `json:"ap"`
		StopPrice           string `json:"sp"`
		ExecutionType       string `json:"x"`
		OrderStatus         string `json:"X"`
		OrderID             int64  `json:"i"`
		LastFilledQuantity  string `json:"l"`
		CumulativeQuantity  string `json:"z"`
		LastFilledPrice     string `json:"L"`
		CommissionAmount    string `json:"n"`
		CommissionAsset     string `json:"N"`
		TradeTime           int64  `json:"T"`
		TradeID             int64  `json:"t"`
		BidNotional         string `json:"b"`
		AskNotional         string `json:"a"`
		IsMaker             bool   `json:"m"`
		IsReduceOnly        bool   `json:"R"`
		WorkingType         string `json:"wt"`
		OriginalOrderType   string `json:"ot"`
		PositionSide        string `json:"ps"`
		IsClosePosition     bool   `json:"cp"`
		RealizedProfit      string `json:"rp"`
		IsPriceProtect      bool   `json:"pP"`
		StopOrderType       int64  `json:"si"`
		StopOrderStatus     int64  `json:"ss"`
		ActivationPriceType string `json:"V"`
		PriceProtectMode    string `json:"pm"`
		GTDTime             int64  `json:"gtd"`
	} `json:"o"`
}
