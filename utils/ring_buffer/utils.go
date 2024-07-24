package ring_buffer

import (
	"math"
)

// Функція для обчислення стандартного відхилення (волатильності)
func CalculateVolatility(prices []float64) float64 {
	var sum, mean, variance float64
	n := float64(len(prices))

	// Обчислення середнього значення
	for _, price := range prices {
		sum += price
	}
	mean = sum / n

	// Обчислення суми квадратів відхилень від середнього
	for _, price := range prices {
		variance += math.Pow(price-mean, 2)
	}

	// Обчислення середнього квадрату відхилень і взяття квадратного кореня
	variance /= n
	return math.Sqrt(variance)
}

// Функція для обчислення кута нахилу тренду з коефіцієнта нахилу a
func CalculateTrendAngle(a float64) float64 {
	return math.Atan(a) * (180 / math.Pi) // Перетворення радіанів в градуси
}

// Функція для додавання нового значення до кільцевого буфера
func AddToBuffer(size int, buffer []float64, value float64) []float64 {
	if len(buffer) < size {
		// Якщо буфер меньше заданого розміру, то додаємо нове значення
		buffer = append(buffer, value)
	} else {
		buffer = append(buffer[1:], value) // Видаляємо перший елемент і додаємо новий в кінець
	}
	return buffer
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
