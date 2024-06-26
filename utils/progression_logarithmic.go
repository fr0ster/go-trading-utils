package utils

import "math"

// LogarithmicProgressionNthTerm обчислює n-й член логарифмічної прогресії.
// Ця функція є нетрадиційною і не є стандартною математичною прогресією.
func LogarithmicProgressionNthTerm(firstTerm, commonRatio float64, termPosition int) float64 {
	// Переконуємося, що позиція члену є дійсною
	if termPosition < 1 {
		return 0 // або обробляємо помилку належним чином
	}
	// Обчислюємо n-й член, використовуючи надану формулу
	return firstTerm + commonRatio*math.Log(float64(termPosition))
}

// FindLogarithmicProgressionNthTerm обчислює n-й член на основі першого та другого членів.
// Ця функція припускає, що другий член слідує тому ж патерну, що й перший, дозволяючи обчислити спільний коефіцієнт.
func FindLogarithmicProgressionNthTerm(firstTerm, secondTerm float64, termPosition int) float64 {
	// Переконуємося, що позиція члену є дійсною
	if termPosition < 1 {
		return 0 // або обробляємо помилку належним чином
	}
	// Обчислюємо спільний коефіцієнт на основі першого та другого членів
	commonRatio := (secondTerm - firstTerm) / math.Log(2)
	// Обчислюємо n-й член, використовуючи отриманий спільний коефіцієнт
	return LogarithmicProgressionNthTerm(firstTerm, commonRatio, termPosition)
}

// LogarithmicProgressionSum обчислює суму перших n членів логарифмічної прогресії.
// Ця функція ітерує кожен член для обчислення суми, оскільки немає прямої формули для суми такої прогресії.
func LogarithmicProgressionSum(firstTerm, commonRatio float64, numberOfTerms int) float64 {
	sum := 0.0
	for i := 1; i <= numberOfTerms; i++ {
		sum += LogarithmicProgressionNthTerm(firstTerm, commonRatio, i)
	}
	return sum
}

// FindLengthOfLogarithmicProgression намагається обчислити довжину логарифмічної прогресії.
// Через природу логарифмічних прогресій, ця операція не має чіткого визначення і тому повертає -1.
func FindLengthOfLogarithmicProgression(firstTerm, secondTerm, lastTerm float64) int {
	// Ця операція не виконується для логарифмічних прогресій, як описано тут.
	return -1
}
