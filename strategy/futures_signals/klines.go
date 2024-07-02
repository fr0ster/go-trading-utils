package futures_signals

import "math"

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

// Функція для обчислення кута нахилу тренду з коефіцієнта нахилу a
func CalculateTrendAngle(a float64) float64 {
	return math.Atan(a) * (180 / math.Pi) // Перетворення радіанів в градуси
}

// Функція для додавання нового значення до кільцевого буфера
func AddToBuffer(buffer []float64, value float64) []float64 {
	buffer = append(buffer[1:], value) // Видаляємо перший елемент і додаємо новий в кінець
	return buffer
}
