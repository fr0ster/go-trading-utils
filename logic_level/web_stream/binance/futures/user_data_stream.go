package futures_api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bitly/go-simplejson"
	types "github.com/fr0ster/go-trading-utils/types"
	api "github.com/fr0ster/turbo-restler/rest_api"
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	common "github.com/fr0ster/turbo-restler/web_stream"
	"github.com/sirupsen/logrus"
)

const (
	UserDataEventTypeListenKeyExpired    types.UserDataEventType = "listenKeyExpired"
	UserDataEventTypeMarginCall          types.UserDataEventType = "MARGIN_CALL"
	UserDataEventTypeAccountUpdate       types.UserDataEventType = "ACCOUNT_UPDATE"
	UserDataEventTypeOrderTradeUpdate    types.UserDataEventType = "ORDER_TRADE_UPDATE"
	UserDataEventTypeAccountConfigUpdate types.UserDataEventType = "ACCOUNT_CONFIG_UPDATE"
)

// WsUserDataEvent define user data event
type WsUserDataEvent struct {
	Event               types.UserDataEventType `json:"e"`
	Time                int64                   `json:"E"`
	CrossWalletBalance  string                  `json:"cw"`
	MarginCallPositions []WsPosition            `json:"p"`
	TransactionTime     int64                   `json:"T"`
	AccountUpdate       WsAccountUpdate         `json:"a"`
	OrderTradeUpdate    WsOrderTradeUpdate      `json:"o"`
	AccountConfigUpdate WsAccountConfigUpdate   `json:"ac"`
}

// WsAccountUpdate define account update
type WsAccountUpdate struct {
	Reason    types.UserDataEventReasonType `json:"m"`
	Balances  []WsBalance                   `json:"B"`
	Positions []WsPosition                  `json:"P"`
}

// WsBalance define balance
type WsBalance struct {
	Asset              string `json:"a"`
	Balance            string `json:"wb"`
	CrossWalletBalance string `json:"cw"`
	ChangeBalance      string `json:"bc"`
}

// WsPosition define position
type WsPosition struct {
	Symbol                    string                 `json:"s"`
	Side                      types.PositionSideType `json:"ps"`
	Amount                    string                 `json:"pa"`
	MarginType                types.MarginType       `json:"mt"`
	IsolatedWallet            string                 `json:"iw"`
	EntryPrice                string                 `json:"ep"`
	MarkPrice                 string                 `json:"mp"`
	UnrealizedPnL             string                 `json:"up"`
	AccumulatedRealized       string                 `json:"cr"`
	MaintenanceMarginRequired string                 `json:"mm"`
}

// WsOrderTradeUpdate define order trade update
type WsOrderTradeUpdate struct {
	Symbol               string                   `json:"s"`   // Symbol
	ClientOrderID        string                   `json:"c"`   // Client order ID
	Side                 types.SideType           `json:"S"`   // Side
	Type                 types.OrderType          `json:"o"`   // Order type
	TimeInForce          types.TimeInForceType    `json:"f"`   // Time in force
	OriginalQty          string                   `json:"q"`   // Original quantity
	OriginalPrice        string                   `json:"p"`   // Original price
	AveragePrice         string                   `json:"ap"`  // Average price
	StopPrice            string                   `json:"sp"`  // Stop price. Please ignore with TRAILING_STOP_MARKET order
	ExecutionType        types.OrderExecutionType `json:"x"`   // Execution type
	Status               types.OrderStatusType    `json:"X"`   // Order status
	ID                   int64                    `json:"i"`   // Order ID
	LastFilledQty        string                   `json:"l"`   // Order Last Filled Quantity
	AccumulatedFilledQty string                   `json:"z"`   // Order Filled Accumulated Quantity
	LastFilledPrice      string                   `json:"L"`   // Last Filled Price
	CommissionAsset      string                   `json:"N"`   // Commission Asset, will not push if no commission
	Commission           string                   `json:"n"`   // Commission, will not push if no commission
	TradeTime            int64                    `json:"T"`   // Order Trade Time
	TradeID              int64                    `json:"t"`   // Trade ID
	BidsNotional         string                   `json:"b"`   // Bids Notional
	AsksNotional         string                   `json:"a"`   // Asks Notional
	IsMaker              bool                     `json:"m"`   // Is this trade the maker side?
	IsReduceOnly         bool                     `json:"R"`   // Is this reduce only
	WorkingType          types.WorkingType        `json:"wt"`  // Stop Price Working Type
	OriginalType         types.OrderType          `json:"ot"`  // Original Order Type
	PositionSide         types.PositionSideType   `json:"ps"`  // Position Side
	IsClosingPosition    bool                     `json:"cp"`  // If Close-All, pushed with conditional order
	ActivationPrice      string                   `json:"AP"`  // Activation Price, only puhed with TRAILING_STOP_MARKET order
	CallbackRate         string                   `json:"cr"`  // Callback Rate, only puhed with TRAILING_STOP_MARKET order
	PriceProtect         bool                     `json:"pP"`  // If price protection is turned on
	RealizedPnL          string                   `json:"rp"`  // Realized Profit of the trade
	STP                  string                   `json:"V"`   // STP mode
	PriceMode            string                   `json:"pm"`  // Price match mode
	GTD                  int64                    `json:"gtd"` // TIF GTD order auto cancel time
}

// WsAccountConfigUpdate define account config update
type WsAccountConfigUpdate struct {
	Symbol   string `json:"s"`
	Leverage int64  `json:"l"`
}

type UserDataStream struct {
	apiKey             string
	sign               signature.Sign
	symbol             string
	websocketKeepalive bool
	useTestNet         bool
	doneC              chan struct{}
	stopC              chan struct{}
}

func (uds *UserDataStream) listenKey(method api.HttpMethod) (listenKey string, err error) {
	const (
		BaseAPIMainUrl    = "https://api.binance.com"
		BaseAPITestnetUrl = "https://testnet.binance.vision"
	)
	baseURL := api.ApiBaseUrl("")
	endpoint := api.EndPoint("/fapi/v1/listenKey")
	if uds.useTestNet {
		baseURL = BaseAPITestnetUrl
	} else {
		baseURL = BaseAPIMainUrl
	}
	var result map[string]interface{}

	body, err := api.CallRestAPI(baseURL, method, nil, endpoint, uds.sign)
	if err != nil {
		return
	}

	// Парсинг відповіді
	err = json.Unmarshal(body, &result)
	listenKey = result["listenKey"].(string)
	return
}

func (uds *UserDataStream) wsHandler(handler func(event *WsUserDataEvent), errHandler func(err error)) func(message *simplejson.Json) {
	return func(message *simplejson.Json) {
		event := new(WsUserDataEvent)
		js, _ := message.MarshalJSON()
		err := json.Unmarshal(js, event)
		if err != nil {
			errHandler(err)
			return
		}
		if event.Event == UserDataEventTypeOrderTradeUpdate && event.OrderTradeUpdate.Symbol != uds.symbol {
			return
		}
		handler(event)
	}
}

func (uds *UserDataStream) Start(callBack func(*WsUserDataEvent)) (err error) {
	var listenKey string
	wss := GetWsBaseUrl(uds.useTestNet)
	listenKey, err = uds.listenKey(http.MethodPost)
	if err != nil {
		return
	}
	wsURL := fmt.Sprintf("%s/%s", wss, listenKey)
	wsErrorHandler := func(err error) {
		logrus.Errorf("error reading from websocket: %v", err)
	}
	uds.doneC, uds.stopC, err = common.StartStreamer(
		wsURL,
		uds.wsHandler(callBack, wsErrorHandler),
		wsErrorHandler,
		uds.websocketKeepalive)
	if err != nil {
		return
	}
	return
}

func NewUserDataStream(apiKey string, symbol string, sign signature.Sign, useTestNet bool, websocketKeepalive ...bool) *UserDataStream {
	var WebsocketKeepalive bool
	if len(websocketKeepalive) > 0 {
		WebsocketKeepalive = websocketKeepalive[0]
	}
	return &UserDataStream{
		apiKey:             apiKey,
		sign:               sign,
		symbol:             symbol,
		websocketKeepalive: WebsocketKeepalive,
		useTestNet:         useTestNet,
		doneC:              make(chan struct{}),
		stopC:              make(chan struct{}),
	}
}
