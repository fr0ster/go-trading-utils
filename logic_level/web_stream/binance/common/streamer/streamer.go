package streamer

import (
	"github.com/bitly/go-simplejson"

	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_stream "github.com/fr0ster/turbo-restler/web_stream"
)

type (
	Streamer struct {
		sign               signature.Sign
		wsURL              string
		websocketKeepalive bool
	}
)

func (st *Streamer) Start(handler web_stream.WsHandler, errHandler web_stream.ErrHandler) (
	doneC chan struct{},
	stopC chan struct{},
	err error) {
	doneC, stopC, err = web_stream.StartStreamer(
		st.wsURL,
		handler,
		errHandler,
		st.websocketKeepalive)
	if err != nil {
		return
	}
	return
}

func New(apiKey, symbol string, url string, sign signature.Sign) *Streamer {
	simpleJson := simplejson.New()
	simpleJson.Set("apiKey", apiKey)
	simpleJson.Set("symbol", symbol)
	return &Streamer{
		sign:  sign,
		wsURL: url,
	}
}
