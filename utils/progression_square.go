package utils

import "math"

// SquareProgressionNthTerm обчислює n-й член квадратичної прогресії.
func SquareProgressionNthTerm(firstTerm, commonRatio float64, termPosition int) float64 {
	return firstTerm * math.Pow(commonRatio, float64(termPosition-1)*2)
}

// FindSquareProgressionNthTerm обчислює n-й член квадратичної прогресії на основі першого та другого членів.
func FindSquareProgressionNthTerm(firstTerm, secondTerm float64, termPosition int) float64 {
	commonRatio := math.Sqrt(secondTerm / firstTerm)
	return firstTerm * math.Pow(commonRatio, float64(termPosition-1)*2)
}

// SquareProgressionSum обчислює суму перших n членів квадратичної прогресії.
func SquareProgressionSum(firstTerm, commonRatio float64, numberOfTerms int) float64 {
	sum := 0.0
	for i := 1; i <= numberOfTerms; i++ {
		sum += firstTerm * math.Pow(commonRatio, float64(i-1)*2)
	}
	return sum
}

// FindLengthOfSquareProgression обчислює кількість членів квадратичної прогресії.
func FindLengthOfSquareProgression(firstTerm, secondTerm, lastTerm float64) int {
	commonRatio := math.Sqrt(secondTerm / firstTerm)
	return int(math.Sqrt(lastTerm/firstTerm)/commonRatio) + 1
}
