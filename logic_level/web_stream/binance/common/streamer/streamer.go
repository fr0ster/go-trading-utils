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
		doneC              chan struct{}
		stopC              chan struct{}
		handler            web_stream.WsHandler
		errHandler         web_stream.ErrHandler
		params             *simplejson.Json
		websocketKeepalive bool
	}
)

func (st *Streamer) Start() (err error) {
	st.doneC, st.stopC, err = web_stream.StartStreamer(
		st.wsURL,
		st.handler,
		st.errHandler,
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
		sign:   sign,
		params: simpleJson,
		wsURL:  url,
	}
}
