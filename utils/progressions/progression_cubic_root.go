package progressions

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

// FindCubicRootProgressionTthTerm calculates the Tth term of a progression
// where each term is obtained by applying cubic root to the previous term,
// given the first term (firstTerm), the last term (lastTerm), the length of the progression (length),
// and the sum of the progression (sum). It returns the value of the Tth term.
func FindCubicRootProgressionTthTerm(firstTerm, lastTerm float64, length int, sum float64, T int) float64 {
	// Correct calculation of the common ratio of the progression
	// Since the progression is defined by cubic roots, the ratio should be calculated accordingly
	commonRatio := math.Pow(lastTerm/firstTerm, 1/float64(length-1))

	// Calculate the Tth term using the corrected formula: TthTerm = firstTerm * commonRatio^(T-1)
	TthTerm := firstTerm * math.Pow(commonRatio, float64(T-1))
	return TthTerm
}
