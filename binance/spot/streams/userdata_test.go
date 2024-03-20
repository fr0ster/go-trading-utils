package streams_test

import (
	"os"
	"testing"

	"github.com/fr0ster/go-trading-utils/binance/spot/client/listenkey"
	"github.com/fr0ster/go-trading-utils/binance/spot/streams"
	"github.com/stretchr/testify/assert"
)

func TestNewUserDataStream(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	listenKey, err := listenkey.New(api_key, secret_key, false).GetListenKey()
	assert.Nil(t, err)
	stream := streams.NewUserDataStream(listenKey)

	if stream == nil {
		t.Error("Expected non-nil UserDataStream, got nil")
	}
}

func TestUserDataStream_Start(t *testing.T) {
	api_key := os.Getenv("API_KEY")
	secret_key := os.Getenv("SECRET_KEY")
	listenKey, err := listenkey.New(api_key, secret_key, false).GetListenKey()
	assert.Nil(t, err)
	stream := streams.NewUserDataStream(listenKey)
	doneC, stopC, err := stream.Start()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if doneC == nil {
		t.Error("Expected non-nil doneC, got nil")
	}

	if stopC == nil {
		t.Error("Expected non-nil stopC, got nil")
	}
}
