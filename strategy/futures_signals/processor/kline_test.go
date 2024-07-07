package processor_test

import (
	"testing"

	processor "github.com/fr0ster/go-trading-utils/strategy/futures_signals/processor"
	"github.com/fr0ster/go-trading-utils/utils"
	"github.com/stretchr/testify/assert"
)

func getTestData() (rb *processor.KlineRingBuffer, klines []float64) {
	rb = processor.NewKlineRingBuffer(5)
	klines = []float64{100, 100, 100, 100, 100}
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

	angle := utils.RoundToDecimalPlace(processor.SlopeToAngle(a), 2)

	assert.Equal(t, expectedAngle, angle)
}
