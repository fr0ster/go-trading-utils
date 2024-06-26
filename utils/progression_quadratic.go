package utils

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
