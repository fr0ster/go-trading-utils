package processor

import (
	"math"
)

type KlineRingBuffer struct {
	elements []float64
	index    int
	size     int
	isFull   bool
}

func NewKlineRingBuffer(size int) *KlineRingBuffer {
	return &KlineRingBuffer{
		elements: make([]float64, size),
		size:     size,
	}
}

func (rb *KlineRingBuffer) Add(element float64) {
	rb.elements[rb.index] = element
	rb.index++
	if rb.index == rb.size {
		rb.index = 0
		rb.isFull = true
	}
}

func (rb *KlineRingBuffer) GetElements() []float64 {
	if !rb.isFull {
		return rb.elements[:rb.index]
	}
	return append(rb.elements[rb.index:], rb.elements[:rb.index]...)
}

func (rb *KlineRingBuffer) GetLastNElements(n int) []float64 {
	elements := rb.GetElements()
	if n > len(elements) {
		return elements
	}
	return elements[len(elements)-n:]
}

func (rb *KlineRingBuffer) GetFirstNElements(n int) []float64 {
	elements := rb.GetElements()
	if n > len(elements) {
		return elements
	}
	return elements[:n]
}

func (rb *KlineRingBuffer) Length() int {
	if rb.isFull {
		return rb.size
	}
	return rb.index
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
func (rb *KlineRingBuffer) FindBestFitLine() (a, b float64) {
	values := make([]float64, len(rb.elements))
	x := make([]float64, len(rb.elements))

	for i, kline := range rb.elements {
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
