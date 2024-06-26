package utils

import "math"

// CubicRootProgressionNthTerm обчислює n-й член прогресії кубічних коренів.
func CubicRootProgressionNthTerm(firstTerm, commonRatio float64, termPosition int) float64 {
	return firstTerm * math.Pow(commonRatio, float64(termPosition-1)/3)
}

// FindCubicRootProgressionNthTerm обчислює n-й член прогресії кубічних коренів на основі першого та другого членів.
func FindCubicRootProgressionNthTerm(firstTerm, secondTerm float64, termPosition int) float64 {
	commonRatio := math.Cbrt(secondTerm / firstTerm)
	return firstTerm * math.Pow(commonRatio, float64(termPosition-1)/3)
}

// CubicRootProgressionSum обчислює суму перших n членів прогресії кубічних коренів.
func CubicRootProgressionSum(firstTerm, commonRatio float64, numberOfTerms int) float64 {
	sum := 0.0
	for i := 1; i <= numberOfTerms; i++ {
		sum += firstTerm * math.Pow(commonRatio, float64(i-1)/3)
	}
	return sum
}

// FindLengthOfCubicRootProgression обчислює кількість членів прогресії кубічних коренів.
func FindLengthOfCubicRootProgression(firstTerm, secondTerm, lastTerm float64) int {
	commonRatio := math.Cbrt(secondTerm / firstTerm)
	return int((3 * math.Log(lastTerm/firstTerm) / math.Log(commonRatio)) + 1)
}
