package progressions

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

// FindLogarithmicProgressionTthTerm calculates the Tth term of a hypothetical logarithmic progression.
// This function assumes a sequence where each term's value is derived from a logarithmic function of its position.
// Parameters:
// - firstTerm: the value of the first term in the sequence.
// - lastTerm: the value of the last term in the sequence.
// - length: the total number of terms in the sequence.
// - sum: the sum of all terms in the sequence (not directly used in this hypothetical model).
// - T: the position of the term to find.
// Returns the value of the Tth term in the sequence.
func FindLogarithmicProgressionTthTerm(firstTerm, lastTerm float64, length int, sum float64, T int) float64 {
	// Hypothetical base for the logarithmic function. This is chosen arbitrarily for this example.
	const base = math.E

	// Calculate the scaling factor (k) to adjust the progression to fit the first and last terms.
	// This is a simplified approach for the sake of this example.
	k := (lastTerm - firstTerm) / (math.Log(float64(length)) / math.Log(base))

	// Calculate the Tth term using the formula: TthTerm = firstTerm + k * log_base(T)
	TthTerm := firstTerm + k*(math.Log(float64(T))/math.Log(base))

	return TthTerm
}
