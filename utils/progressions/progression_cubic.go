package progressions

import "math"

// CubicProgressionNthTerm обчислює n-й член кубічної прогресії.
func CubicProgressionNthTerm(firstTerm, commonRatio float64, termPosition int) float64 {
	return firstTerm * math.Pow(commonRatio, float64(termPosition-1)*3)
}

// FindCubicProgressionNthTerm обчислює n-й член кубічної прогресії на основі першого та другого членів.
func FindCubicProgressionNthTerm(firstTerm, secondTerm float64, termPosition int) float64 {
	commonRatio := math.Pow(secondTerm/firstTerm, 1.0/3)
	return firstTerm * math.Pow(commonRatio, float64(termPosition-1)*3)
}

// CubicProgressionSum обчислює суму перших n членів кубічної прогресії.
func CubicProgressionSum(firstTerm, commonRatio float64, numberOfTerms int) float64 {
	sum := 0.0
	for i := 1; i <= numberOfTerms; i++ {
		sum += firstTerm * math.Pow(commonRatio, float64(i-1)*3)
	}
	return sum
}

// FindLengthOfCubicProgression обчислює кількість членів кубічної прогресії.
func FindLengthOfCubicProgression(firstTerm, secondTerm, lastTerm float64) int {
	commonRatio := math.Pow(secondTerm/firstTerm, 1.0/3)
	return int(math.Cbrt(lastTerm/firstTerm)/commonRatio) + 1
}

// FindCubicProgressionTthTerm calculates the Tth term of a cubic progression
// given the first term (firstTerm), the last term (lastTerm), the length of the progression (length),
// and the sum of the progression (sum). It returns the value of the Tth term in the cubic progression.
func FindCubicProgressionTthTerm(firstTerm, lastTerm float64, length int, sum float64, T int) float64 {
	// Calculate the common ratio of the cubic progression
	commonRatio := math.Pow(lastTerm/firstTerm, 1/float64((length-1)/3))
	// Calculate the Tth term using the formula: TthTerm = firstTerm * commonRatio^(3*(T-1))
	TthTerm := firstTerm * math.Pow(commonRatio, float64(3*(T-1)))
	return TthTerm
}
