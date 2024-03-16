package types

type ExchangeInfo struct {
	Timezone   string       `json:"timezone"`
	ServerTime int64        `json:"serverTime"`
	RateLimits []RateLimit  `json:"rateLimits"`
	Symbols    []SymbolInfo `json:"symbols"`
}

type RateLimit struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	Limit         int    `json:"limit"`
}

type SymbolInfo struct {
	Symbol                 string         `json:"symbol"`
	Status                 string         `json:"status"`
	BaseAsset              string         `json:"baseAsset"`
	BaseAssetPrecision     int            `json:"baseAssetPrecision"`
	QuoteAsset             string         `json:"quoteAsset"`
	QuotePrecision         int            `json:"quotePrecision"`
	Filters                []SymbolFilter `json:"filters"`
	OrderTypes             []string       `json:"orderTypes"`
	TimeInForceTypes       []string       `json:"timeInForceTypes"`
	IcebergAllowed         bool           `json:"icebergAllowed"`
	OcoAllowed             bool           `json:"ocoAllowed"`
	QuoteOrderQtyScale     int            `json:"quoteOrderQtyScale"`
	MinPrice               string         `json:"minPrice"`
	MaxPrice               string         `json:"maxPrice"`
	MinQty                 string         `json:"minQty"`
	MaxQty                 string         `json:"maxQty"`
	StepSize               string         `json:"stepSize"`
	MinNotional            string         `json:"minNotional"`
	IsSpotTradingAllowed   bool           `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed bool           `json:"isMarginTradingAllowed"`
}

type SymbolFilter struct {
	FilterType  string `json:"filterType"`
	MinPrice    string `json:"minPrice,omitempty"`
	MaxPrice    string `json:"maxPrice,omitempty"`
	TickSize    string `json:"tickSize,omitempty"`
	MinQty      string `json:"minQty,omitempty"`
	MaxQty      string `json:"maxQty,omitempty"`
	StepSize    string `json:"stepSize,omitempty"`
	MinNotional string `json:"minNotional,omitempty"`
}
