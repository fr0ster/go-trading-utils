package utils

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
