package streamer

import (
	web_api "github.com/fr0ster/turbo-restler/web_api"
	web_stream "github.com/fr0ster/turbo-restler/web_stream"
)

type (
	Stream struct {
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

func New(host web_api.WsHost, path web_api.WsPath) *Stream {
	return &Stream{
		wsHost: host,
		wsPath: path,
	}
}
