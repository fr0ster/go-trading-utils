package futures_api

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// var dialer = websocket.DefaultDialer

func StartStreamer(url string, callBack func([]byte), quit chan struct{}) {
	wsServe(url, callBack, func(err error) { logrus.Fatalf("Error reading from websocket: %v", err) })
	// conn, _, err := dialer.Dial(url, nil)
	// if err != nil {
	// 	logrus.Fatalf("dial: %v", err)
	// }
	// defer conn.Close()

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// go func() {
	// 	go func() {
	// 		<-ctx.Done()
	// 		// Закриваємо з'єднання з сервером
	// 		err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 		if err != nil {
	// 			logrus.Infof("write close: %v", err)
	// 			return
	// 		}
	// 		cancel()
	// 	}()
	// 	for {
	// 		select {
	// 		case <-quit:
	// 			cancel()
	// 		// case <-ctx.Done():
	// 		// 	// Закриваємо з'єднання з сервером
	// 		// 	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	// 		// 	if err != nil {
	// 		// 		logrus.Infof("write close: %v", err)
	// 		// 		return
	// 		// 	}
	// 		// 	return
	// 		default:
	// 			_, message, err := conn.ReadMessage()
	// 			if err != nil {
	// 				return
	// 			}
	// 			callBack(message)
	// 			// time.Sleep(1000 * time.Microsecond)
	// 		}
	// 	}
	// }()

	// <-quit
	// // cancel()
}

var (
	// // Endpoints
	// BaseWsMainURL          = "wss://stream.binance.com:9443/ws"
	// BaseWsTestnetURL       = "wss://testnet.binance.vision/ws"
	// BaseCombinedMainURL    = "wss://stream.binance.com:9443/stream?streams="
	// BaseCombinedTestnetURL = "wss://testnet.binance.vision/stream?streams="

	// WebsocketTimeout is an interval for sending ping/pong messages if WebsocketKeepalive is enabled
	WebsocketTimeout = time.Second * 60
	// WebsocketKeepalive enables sending ping/pong messages to check the connection stability
	WebsocketKeepalive = false
)

// WsHandler handle raw websocket message
type WsHandler func(message []byte)

// ErrHandler handles errors
type ErrHandler func(err error)

func wsServe(endpoint string, handler WsHandler, errHandler ErrHandler) (doneC, stopC chan struct{}, err error) {
	Dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: false,
	}

	c, _, err := Dialer.Dial(endpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	c.SetReadLimit(655350)
	doneC = make(chan struct{})
	stopC = make(chan struct{})
	go func() {
		// This function will exit either on error from
		// websocket.Conn.ReadMessage or when the stopC channel is
		// closed by the client.
		defer close(doneC)
		if WebsocketKeepalive {
			keepAlive(c, WebsocketTimeout)
		}
		// Wait for the stopC channel to be closed.  We do that in a
		// separate goroutine because ReadMessage is a blocking
		// operation.
		silent := false
		go func() {
			select {
			case <-stopC:
				silent = true
			case <-doneC:
			}
			c.Close()
		}()
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				if !silent {
					errHandler(err)
				}
				return
			}
			handler(message)
		}
	}()
	return
}

func keepAlive(c *websocket.Conn, timeout time.Duration) {
	ticker := time.NewTicker(timeout)

	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		defer ticker.Stop()
		for {
			deadline := time.Now().Add(10 * time.Second)
			err := c.WriteControl(websocket.PingMessage, []byte{}, deadline)
			if err != nil {
				return
			}
			<-ticker.C
			if time.Since(lastResponse) > timeout {
				c.Close()
				return
			}
		}
	}()
}
