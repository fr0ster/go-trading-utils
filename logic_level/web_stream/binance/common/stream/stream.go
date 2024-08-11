package streamer

import (
	signature "github.com/fr0ster/turbo-restler/utils/signature"
	web_api "github.com/fr0ster/turbo-restler/web_api"
	web_stream "github.com/fr0ster/turbo-restler/web_stream"
)

type (
	Stream struct {
		sign               signature.Sign
		wsHost             web_api.WsHost
		wsPath             web_api.WsPath
		websocketKeepalive bool
	}
)

func (rq *Stream) Start(handler web_stream.WsHandler, errHandler web_stream.ErrHandler) (
	doneC chan struct{},
	stopC chan struct{},
	err error) {
	doneC, stopC, err = web_stream.StartStreamer(
		rq.wsHost,
		rq.wsPath,
		handler,
		errHandler,
		rq.websocketKeepalive)
	if err != nil {
		return
	}
	return
}

func New(symbol string, host web_api.WsHost, path web_api.WsPath, sign signature.Sign) *Stream {
	return &Stream{
		sign:   sign,
		wsHost: host,
		wsPath: path,
	}
}
