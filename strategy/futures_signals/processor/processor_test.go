package processor_test

import (
	"testing"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"

	"github.com/stretchr/testify/assert"
)

const msg = "Way too many requests; IP(88.238.33.41) banned until 1721393459728. Please use the websocket for live updates to avoid bans."

func TestParserStr(t *testing.T) {
	ip, time, err := processor.ParserErr1003(msg)
	assert.Nil(t, err)
	assert.Equal(t, "88.238.33.41", ip)
	assert.Equal(t, "1721393459728", time)
}
