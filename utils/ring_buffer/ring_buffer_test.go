package ring_buffer_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/utils"
	ring_buffer "github.com/fr0ster/go-trading-utils/utils/ring_buffer"
	"github.com/stretchr/testify/assert"
)

func getTestData(klines []float64) (rb *ring_buffer.RingBuffer) {
	rb = ring_buffer.NewRingBuffer(5, 5)
	for _, kline := range klines {
		rb.Add(kline)
	}
	return
}

func TestKlineRingBuffer_Add(t *testing.T) {
	expectedElements := []float64{100, 100, 100, 100, 100}
	rb := getTestData(expectedElements)

	elements := rb.GetElements()

	if len(elements) != 5 {
		t.Errorf("Expected ring buffer size to be 5, got %d", len(elements))
	}
	for i, expected := range expectedElements {
		if elements[i] != expected {
			t.Errorf("Expected element at index %d to be %v, got %v", i, expected, elements[i])
		}
	}
	rb = ring_buffer.NewRingBuffer(5, 5)
	rb.Add(100)
	rb.Add(100)
	rb.Add(100)
	rb.Add(100)
	rb.Add(100)

	rb.Add(110)
	rb.Add(120)
	rb.Add(130)
	rb.Add(140)
	rb.Add(150)
	rb.Add(160)
	elements = rb.GetElements()
	expectedElements = []float64{120, 130, 140, 150, 160}
	if len(elements) != 5 {
		t.Errorf("Expected ring buffer size to be 5, got %d", len(elements))
	}
	for i, expected := range expectedElements {
		if elements[i] != expected {
			t.Errorf("Expected element at index %d to be %v, got %v", i, expected, elements[i])
		}
	}
	assert.Equal(t, 700.0, rb.Summa())
}

func TestKlineRingBuffer_FindBestFitLine(t *testing.T) {
	expectedElements := []float64{100, 100, 100, 100, 100}
	rb := getTestData(expectedElements)
	a, b := ring_buffer.FindBestFitLine(rb.GetElements())
	assert.Equal(t, 0.0, a)
	assert.Equal(t, 100.0, b)

	expectedElements = []float64{100, 101, 102, 103, 104}
	rb = getTestData(expectedElements)
	a, b = ring_buffer.FindBestFitLine(rb.GetElements())
	assert.Equal(t, 1.0, a)
	assert.Equal(t, 100.0, b)

	expectedElements = []float64{57923.8, 57407.2, 57935.9, 57818.7, 57822.5, 57417.5, 57742.3, 57558.5, 57620, 57472.5}
	rb = getTestData(expectedElements)
	a, b = ring_buffer.FindBestFitLine(rb.GetElements())
	assert.Equal(t, -1.23, utils.RoundToDecimalPlace(a, 2))
	assert.Equal(t, 57564.619999999995, b)
}

func TestSlopeToAngle(t *testing.T) {
	a := 1.5
	expectedAngle := 56.31
	angle := utils.RoundToDecimalPlace(ring_buffer.SlopeToAngle(a), 2)
	assert.Equal(t, expectedAngle, angle)

	a = 0
	expectedAngle = 0
	angle = utils.RoundToDecimalPlace(ring_buffer.SlopeToAngle(a), 2)
	assert.Equal(t, expectedAngle, angle)

	a = 10
	expectedAngle = 84.29
	angle = utils.RoundToDecimalPlace(ring_buffer.SlopeToAngle(a), 2)
	assert.Equal(t, expectedAngle, angle)

	a = -1.23
	expectedAngle = -50.89
	angle = utils.RoundToDecimalPlace(ring_buffer.SlopeToAngle(a), 2)
	assert.Equal(t, expectedAngle, angle)
}

func TestKlineRingBuffer_GetLastNElements(t *testing.T) {
	expectedElements := []float64{100, 101, 102, 103, 104}
	rb := getTestData(expectedElements)
	lastNElements := rb.GetLastNElements(3)
	assert.Equal(t, []float64{102, 103, 104}, lastNElements)

	lastNElements = rb.GetLastNElements(5)
	assert.Equal(t, expectedElements, lastNElements)

	lastNElements = rb.GetLastNElements(6)
	assert.Equal(t, expectedElements, lastNElements)
}

func TestKlineRingBuffer_GetFirstNElements(t *testing.T) {
	expectedElements := []float64{100, 101, 102, 103, 104}
	rb := getTestData(expectedElements)
	firstNElements := rb.GetFirstNElements(3)
	assert.Equal(t, []float64{100, 101, 102}, firstNElements)

	firstNElements = rb.GetFirstNElements(5)
	assert.Equal(t, expectedElements, firstNElements)

	firstNElements = rb.GetFirstNElements(6)
	assert.Equal(t, expectedElements, firstNElements)
}

func TestKlineRingBuffer_GetElements(t *testing.T) {
	expectedElements := []float64{100, 101, 102, 103, 104}
	rb := getTestData(expectedElements)
	elements := rb.GetElements()
	assert.Equal(t, expectedElements, elements)
}

func TestKlineRingBuffer_GetElementsPercentageChange(t *testing.T) {
	expectedElements := []float64{100, 101, 102, 103, 104}
	rb := getTestData(expectedElements)
	percentageChange := rb.GetElementsPercentageChange()
	assert.Equal(t, []float64{100, 101, 102, 103, 104}, percentageChange)

	expectedElements = []float64{100, 110, 120, 130, 140}
	rb = getTestData(expectedElements)
	percentageChange = rb.GetElementsPercentageChange()
	assert.Equal(t, []float64{100, 110, 120, 130, 140}, percentageChange)

	expectedElements = []float64{100, 90, 80, 70, 60}
	rb = getTestData(expectedElements)
	percentageChange = rb.GetElementsPercentageChange()
	assert.Equal(t, []float64{100, 90, 80, 70, 60}, percentageChange)
}
