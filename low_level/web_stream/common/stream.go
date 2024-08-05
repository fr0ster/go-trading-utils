package futures_api

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var dialer = websocket.DefaultDialer

func StartStreamer(url string, callBack func([]byte), quit chan struct{}) {
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		logrus.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		go func() {
			<-ctx.Done()
			// Закриваємо з'єднання з сервером
			err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logrus.Infof("write close: %v", err)
				return
			}
			cancel()
		}()
		for {
			select {
			case <-quit:
				cancel()
			// case <-ctx.Done():
			// 	// Закриваємо з'єднання з сервером
			// 	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			// 	if err != nil {
			// 		logrus.Infof("write close: %v", err)
			// 		return
			// 	}
			// 	return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					return
				}
				callBack(message)
				// time.Sleep(1000 * time.Microsecond)
			}
		}
	}()

	<-quit
	// cancel()
}
