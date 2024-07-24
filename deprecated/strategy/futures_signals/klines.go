package futures_signals

import (
	"math"
)

// Функція для обчислення коефіцієнтів a та b за методом найменших квадратів
func LeastSquares(y []float64) (a float64, b float64) {
	var sumX, sumY, sumXY, sumX2 float64
	x := []float64{}
	// Генерувати динамічно в залежності від довжини буфера
	for i := 0; i < len(y); i++ {
		x = append(x, float64(i))
	}
	N := float64(len(x))

	for i := 0; i < int(N); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
	}

	a = (N*sumXY - sumX*sumY) / (N*sumX2 - sumX*sumX)
	b = (sumY - a*sumX) / N

	return a, b
}

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
