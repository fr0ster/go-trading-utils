package processor_test

import (
	"testing"

	"github.com/adshao/go-binance/v2/futures"
	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	"github.com/stretchr/testify/assert"
)

func getTestData() (rb *processor.KlineRingBuffer, klines []*futures.Kline) {
	rb = processor.NewKlineRingBuffer(5)
	klines = []*futures.Kline{
		{Open: "100", Close: "100"},
		{Open: "100", Close: "100"},
		{Open: "100", Close: "100"},
		{Open: "100", Close: "100"},
		{Open: "100", Close: "100"},
	}
	for _, kline := range klines {
		rb.Add(kline)
	}
	return
}

func TestKlineRingBuffer_Add(t *testing.T) {
	rb, expectedElements := getTestData()

	elements := rb.GetElements()

	if len(elements) != 5 {
		t.Errorf("Expected ring buffer size to be 5, got %d", len(elements))
	}

	for i, expected := range expectedElements {
		if elements[i] != expected {
			t.Errorf("Expected element at index %d to be %v, got %v", i, expected, elements[i])
		}
	}
}

func TestKlineRingBuffer_FindBestFitLine(t *testing.T) {
	rb, _ := getTestData()

	a, b := rb.FindBestFitLine()

	// Add your assertions here to test the values of a and b

	assert.Equal(t, 0.0, a)
	assert.Equal(t, 100.0, b)
}

func TestSlopeToAngle(t *testing.T) {
	a := 1.5
	expectedAngle := 56.31

	angle := processor.SlopeToAngle(a)

	if angle != expectedAngle {
		t.Errorf("Expected angle to be %.2f, got %.2f", expectedAngle, angle)
	}
}
