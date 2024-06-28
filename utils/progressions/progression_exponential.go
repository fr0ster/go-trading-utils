package progressions

import "math"

// ExponentialProgressionNthTerm обчислює n-й член експоненційної прогресії
// на основі першого члена (firstTerm), спільного відношення (commonRatio) та позиції члена (termPosition).
// Функція повертає значення n-го члена експоненційної прогресії.
func ExponentialProgressionNthTerm(firstTerm, commonRatio float64, termPosition int) float64 {
	return firstTerm * math.Pow(commonRatio, float64(termPosition-1))
}

// FindExponentialProgressionNthTerm обчислює n-й член експоненційної прогресії
// на основі першого та другого членів (firstTerm, secondTerm) та позиції члена (termPosition).
// Функція повертає значення n-го члена експоненційної прогресії.
func FindExponentialProgressionNthTerm(firstTerm, secondTerm float64, termPosition int) float64 {
	commonRatio := secondTerm / firstTerm
	return firstTerm * math.Pow(commonRatio, float64(termPosition-1))
}

// ExponentialProgressionSum обчислює суму перших n членів експоненційної прогресії.
// Вона приймає перший член (firstTerm), спільне відношення (commonRatio), та кількість членів (numberOfTerms) як вхідні дані.
// Функція повертає суму перших n членів експоненційної прогресії.
func ExponentialProgressionSum(firstTerm, commonRatio float64, numberOfTerms int) float64 {
	if commonRatio == 1 {
		return firstTerm * float64(numberOfTerms)
	}
	return firstTerm * (1 - math.Pow(commonRatio, float64(numberOfTerms))) / (1 - commonRatio)
}

// FindLengthOfExponentialProgression обчислює кількість членів експоненційної прогресії
// на основі першого члена (firstTerm), другого члена (secondTerm), та останнього члена (lastTerm).
// Функція повертає кількість членів експоненційної прогресії.
func FindLengthOfExponentialProgression(firstTerm, secondTerm, lastTerm float64) int {
	commonRatio := secondTerm / firstTerm
	return int(math.Log(lastTerm/firstTerm)/math.Log(commonRatio)) + 1
}

// FindExponentialProgressionTthTerm calculates the Tth term of an exponential progression
// given the first term (a), the last term (l), the length of the progression (n),
// and the sum of the progression (S). It returns the value of the Tth term (TthTerm).
func FindExponentialProgressionTthTerm(a, l float64, n int, S float64, T int) float64 {
	// Calculate the common ratio (r) using the formula: r = (l/a)^(1/(n-1))
	r := math.Pow(l/a, 1/float64(n-1))

	// Calculate the Tth term using the formula: TthTerm = a * r^(T-1)
	TthTerm := a * math.Pow(r, float64(T-1))

	return TthTerm
}
