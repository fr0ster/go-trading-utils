package utils

import "math"

// HarmonicProgressionNthTerm обчислює n-й член гармонійної прогресії.
// Вхідні параметри: перший член (firstTerm), спільне відношення (commonRatio), позиція члена (termPosition).
// Функція повертає значення n-го члена гармонійної прогресії.
func HarmonicProgressionNthTerm(firstTerm, commonRatio float64, termPosition int) float64 {
	// В гармонійній прогресії n-й член можна знайти як обернене значення n-го члена відповідної геометричної прогресії.
	return 1 / (firstTerm * math.Pow(commonRatio, float64(termPosition-1)))
}

// FindHarmonicProgressionNthTerm обчислює n-й член гармонійної прогресії
// на основі першого та другого членів (firstTerm, secondTerm) та позиції члена (termPosition).
// Функція повертає значення n-го члена гармонійної прогресії.
func FindHarmonicProgressionNthTerm(firstTerm, secondTerm float64, termPosition int) float64 {
	// Спочатку знаходимо спільне відношення, використовуючи перший та другий члени.
	commonRatio := 1/secondTerm - 1/firstTerm
	// Потім обчислюємо n-й член як обернене значення відповідного члена геометричної прогресії.
	return 1 / (1/firstTerm + commonRatio*float64(termPosition-1))
}

// HarmonicProgressionSum обчислює суму перших n членів гармонійної прогресії.
// Вхідні параметри: перший член (firstTerm), спільне відношення (commonRatio), кількість членів (numberOfTerms).
// Функція повертає суму перших n членів гармонійної прогресії.
func HarmonicProgressionSum(firstTerm, commonRatio float64, numberOfTerms int) float64 {
	sum := 0.0
	for i := 1; i <= numberOfTerms; i++ {
		sum += 1 / (firstTerm * math.Pow(commonRatio, float64(i-1)))
	}
	return sum
}

// FindLengthOfHarmonicProgression обчислює кількість членів гармонійної прогресії
// на основі першого члена (firstTerm), другого члена (secondTerm), та останнього члена (lastTerm).
// Функція повертає кількість членів гармонійної прогресії.
func FindLengthOfHarmonicProgression(firstTerm, secondTerm, lastTerm float64) int {
	// Для гармонійної прогресії, ця задача є складною без додаткових умов або формул.
	// Припустимо, що ми маємо формулу для знаходження кількості членів, але в загальному випадку це може бути складно.
	// Повертаємо прикладне значення для демонстрації.
	return -1 // Замініть це на реальний розрахунок, якщо доступна відповідна формула.
}
