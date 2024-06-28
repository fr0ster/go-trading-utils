package progressions

import "math"

// QuadraticProgressionNthTerm обчислює n-й член квадратичної прогресії.
// Вхідні параметри: перший член (firstTerm), спільне відношення (commonRatio), позиція члена (termPosition).
// Функція повертає значення n-го члена квадратичної прогресії.
func QuadraticProgressionNthTerm(firstTerm, commonDifference float64, termPosition int) float64 {
	return firstTerm + commonDifference*float64((termPosition-1)*(termPosition-1))
}

// FindQuadraticProgressionNthTerm обчислює n-й член квадратичної прогресії
// на основі першого та другого членів (firstTerm, secondTerm) та позиції члена (termPosition).
// Функція повертає значення n-го члена квадратичної прогресії.
func FindQuadraticProgressionNthTerm(firstTerm, secondTerm float64, termPosition int) float64 {
	commonDifference := (secondTerm - firstTerm) / 1 // Для квадратичної прогресії, це не просто різниця.
	return firstTerm + commonDifference*float64((termPosition-1)*(termPosition-1))
}

// QuadraticProgressionSum обчислює суму перших n членів квадратичної прогресії.
// Вхідні параметри: перший член (firstTerm), спільне відношення (commonRatio), кількість членів (numberOfTerms).
// Функція повертає суму перших n членів квадратичної прогресії.
func QuadraticProgressionSum(firstTerm, commonDifference float64, numberOfTerms int) float64 {
	sum := 0.0
	for i := 1; i <= numberOfTerms; i++ {
		sum += firstTerm + commonDifference*float64((i-1)*(i-1))
	}
	return sum
}

// FindLengthOfQuadraticProgression обчислює кількість членів квадратичної прогресії
// на основі першого члена (firstTerm), другого члена (secondTerm), та останнього члена (lastTerm).
// Функція повертає кількість членів квадратичної прогресії.
// Зауваження: Ця функція може бути складною для реалізації без конкретної формули для квадратичної прогресії,
// тому тут наведено прикладний підхід, який може потребувати додаткової адаптації.
func FindLengthOfQuadraticProgression(firstTerm, secondTerm, lastTerm float64) int {
	// Ця задача вимагає розв'язання квадратного рівняння для знаходження кількості членів,
	// що може бути нетривіальним без конкретної формули квадратичної прогресії.
	// Припустимо, що commonDifference = (secondTerm - firstTerm) / 1 для спрощення.
	commonDifference := (secondTerm - firstTerm) / 1
	// Розв'язок квадратного рівняння може бути реалізований тут.
	// Повертаємо прикладне значення для демонстрації.
	return int(math.Sqrt((lastTerm-firstTerm)/commonDifference)) + 1
}

// FindQuadraticProgressionTthTerm calculates the Tth term of a quadratic progression
// given the first term (a), the last term (l), the length of the progression (n),
// and the sum of the progression (S). It returns the value of the Tth term (TthTerm).
func FindQuadraticProgressionTthTerm(a, l float64, n int, S float64, T int) float64 {
	// Calculate the common difference (d) using the formula: d = (2S/n - 2a - 2l + 2a/n + 2l/n) / (n - 1)
	d := (2*S/float64(n) - 2*a - 2*l + 2*a/float64(n) + 2*l/float64(n)) / float64(n-1)

	// // Calculate the second term (b) of the progression for further calculations
	// b := a + d

	// Calculate the Tth term using the formula: TthTerm = a + (T-1)*d + (T-1)*(T-2)*d^2/(2*n)
	TthTerm := a + float64(T-1)*d + (float64(T-1)*(float64(T-2)*math.Pow(d, 2)))/(2*float64(n))

	return TthTerm
}
