package accounts

import "time"

// Exchange identifiers
const (
	ExchangeBinance = "binance"
	ExchangeKraken  = "kraken"
	ExchangeBybit   = "bybit"
)

// Market types
const (
	MarketSpot    = "spot"
	MarketFutures = "futures"
)

// Common metadata for account events
// Version is a monotonically increasing integer per account stream
// Ts is the event time in UTC
// Source hints origin (ws/rest/recon)
type Meta struct {
	TenantID  string    `json:"tenant_id"`
	AccountID string    `json:"account_id"`
	Exchange  string    `json:"exchange"`
	Market    string    `json:"market"`
	Version   uint64    `json:"version"`
	Ts        time.Time `json:"ts"`
	Source    string    `json:"source"`
}

// BalanceUpdateEvent reflects change in balance or margin
// Free = available, Locked = in orders
// For futures, use Wallet/UnrealizedPnL fields as needed
// Zero values can be omitted by omitempty
type BalanceUpdateEvent struct {
	Meta   Meta    `json:"meta"`
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`
	Locked float64 `json:"locked,omitempty"`
	Wallet float64 `json:"wallet,omitempty"`
	UPnL   float64 `json:"upnl,omitempty"`
}

// PositionUpdateEvent for futures positions
type PositionUpdateEvent struct {
	Meta     Meta    `json:"meta"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"` // LONG/SHORT/FLAT
	Quantity float64 `json:"qty"`
	EntryPx  float64 `json:"entry_px"`
	MarkPx   float64 `json:"mark_px"`
	Leverage float64 `json:"leverage,omitempty"`
	UPnL     float64 `json:"upnl"`
	ROE      float64 `json:"roe,omitempty"`
}

// Order statuses
const (
	OrderNew        = "NEW"
	OrderPartFilled = "PARTIALLY_FILLED"
	OrderFilled     = "FILLED"
	OrderCanceled   = "CANCELED"
	OrderExpired    = "EXPIRED"
	OrderRejected   = "REJECTED"
)

// OrderUpdateEvent captures order lifecycle
type OrderUpdateEvent struct {
	Meta      Meta    `json:"meta"`
	OrderID   string  `json:"order_id"`
	ClientOID string  `json:"client_oid,omitempty"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"` // BUY/SELL
	Type      string  `json:"type"` // LIMIT/MARKET/STOP/TAKE_PROFIT/etc
	Status    string  `json:"status"`
	Price     float64 `json:"price,omitempty"`
	StopPrice float64 `json:"stop_price,omitempty"`
	Qty       float64 `json:"qty"`
	FilledQty float64 `json:"filled_qty"`
	AvgFillPx float64 `json:"avg_fill_px,omitempty"`
}

// AccountSnapshotEvent provides a complete view for reconciliation
type AccountSnapshotEvent struct {
	Meta      Meta                  `json:"meta"`
	Balances  []BalanceUpdateEvent  `json:"balances"`
	Positions []PositionUpdateEvent `json:"positions,omitempty"`
	Orders    []OrderUpdateEvent    `json:"orders"`
}
