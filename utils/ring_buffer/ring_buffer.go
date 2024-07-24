package ring_buffer

import (
	"math"

	"github.com/fr0ster/go-trading-utils/utils"
)

type RingBuffer struct {
	elements                 []float64
	elementsPercentageChange []float64
	first                    int
	last                     int
	size                     int
	summa                    float64
	threshold                float64
	isFull                   bool
}

func NewRingBuffer(size int, threshold float64) *RingBuffer {
	return &RingBuffer{
		elements:                 make([]float64, size),
		elementsPercentageChange: make([]float64, size),
		size:                     size,
		threshold:                threshold,
	}
}

func (rb *RingBuffer) Add(element float64) {
	// Додаємо елемент та оновлюємо індекс останнього елемента
	rb.summa += element - rb.elements[rb.last]
	rb.elements[rb.last] = element
	rb.elementsPercentageChange[rb.last] = utils.RoundToDecimalPlace(element/rb.elements[rb.first]*100, 6)
	rb.last = (rb.last + 1) % rb.size
	// Якщо буфер заповнений, зсуваємо індекс першого елемента
	if rb.isFull {
		rb.first = (rb.last + 1) % rb.size
	}
	// Перевіряємо, чи буфер став заповненим
	if rb.last == rb.size-1 {
		rb.isFull = true
	}
}

func (rb *RingBuffer) GetElements() []float64 {
	if !rb.isFull {
		return rb.elements[:rb.last]
	}
	return append(rb.elements[rb.last:], rb.elements[:rb.last]...)
}

func (rb *RingBuffer) SetElements(elements []float64) {
	for _, element := range elements {
		rb.Add(element)
	}
}

func (rb *RingBuffer) GetLastNElements(n int) []float64 {
	if !rb.isFull {
		return []float64{}
	} else if n <= len(rb.GetElements()) {
		return rb.GetElements()[len(rb.GetElements())-n:]
	} else {
		return rb.GetElements()
	}
}

func (rb *RingBuffer) GetFirstNElements(n int) []float64 {
	if !rb.isFull {
		return []float64{}
	} else if n <= len(rb.GetElements()) {
		return rb.GetElements()[:n]
	} else {
		return rb.GetElements()
	}
}

func (rb *RingBuffer) GetElementsPercentageChange() []float64 {
	if !rb.isFull {
		return rb.elementsPercentageChange[:rb.last]
	}
	return append(rb.elementsPercentageChange[rb.last:], rb.elementsPercentageChange[:rb.last]...)
}

func (rb *RingBuffer) GetLastNElementsPercentageChange(n int) []float64 {
	elements := rb.GetLastNElements(n)
	percentageChange := make([]float64, len(elements))
	percentageChange[0] = 100
	for i := 1; i < len(elements); i++ {
		percentageChange[i] = elements[i] / elements[0] * 100
	}
	return percentageChange
}

func (rb *RingBuffer) GetFirstNElementsPercentageChange(n int) []float64 {
	elements := rb.GetFirstNElements(n)
	percentageChange := make([]float64, len(elements))
	percentageChange[0] = 100
	for i := 1; i < len(elements); i++ {
		percentageChange[i] = elements[i] / elements[0] * 100
	}
	return percentageChange
}

func (rb *RingBuffer) Length() int {
	if rb.isFull {
		return rb.size
	}
	return rb.last
}

func (rb *RingBuffer) Summa() float64 {
	return rb.summa
}

func (rb *RingBuffer) GetTrend() (a, b, angle float64) {
	new := rb.GetElementsPercentageChange()
	a, b = FindBestFitLine(new)
	angle = SlopeToAngle(a)
	return
}

func (rb *RingBuffer) GetSlope() float64 {
	new := rb.GetElementsPercentageChange()
	a, _ := FindBestFitLine(new)
	return a
}

func (rb *RingBuffer) GetIntercept() float64 {
	new := rb.GetElementsPercentageChange()
	_, b := FindBestFitLine(new)
	return b
}

func (rb *RingBuffer) GetAngle() float64 {
	_, _, angle := rb.GetTrend()
	return angle
}

func (rb *RingBuffer) IsUp() bool {
	_, _, angle := rb.GetTrend()
	return angle > rb.threshold
}

func (rb *RingBuffer) IsDown() bool {
	_, _, angle := rb.GetTrend()
	return angle < -rb.threshold
}

func (rb *RingBuffer) IsFlat() bool {
	_, _, angle := rb.GetTrend()
	return math.Abs(angle) < rb.threshold
}

func (rb *RingBuffer) IsFull() bool {
	return rb.isFull
}

func (rb *RingBuffer) GetThreshold() float64 {
	return rb.threshold
}

func (rb *RingBuffer) SetThreshold(threshold float64) {
	rb.threshold = threshold
}

// Функція для розрахунку коефіцієнтів прямої методом найменших квадратів
func LeastSquares(x, y []float64) (a, b float64) {
	var sumX, sumY, sumXY, sumXX float64
	n := float64(len(x))

	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumXX += x[i] * x[i]
	}

	a = (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	b = (sumY - a*sumX) / n

	return a, b
}

// Функція для знаходження прямої, яка найменше відхиляється від N останніх найбільших значень close і open
func FindBestFitLine(rb []float64) (a, b float64) {
	values := make([]float64, len(rb))
	x := make([]float64, len(rb))

	for i, kline := range rb {
		values[i] = kline
		x[i] = float64(i)
	}

	a, b = LeastSquares(x, values)
	return a, b
}

// Припустимо, що a - це коефіцієнт нахилу, отриманий з FindBestFitLine
func SlopeToAngle(a float64) float64 {
	angleRadians := math.Atan(a)                   // Переводимо нахил в радіани
	angleDegrees := angleRadians * (180 / math.Pi) // Переводимо радіани в градуси
	return angleDegrees
}
