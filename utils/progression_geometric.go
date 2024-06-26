package utils

import "math"

// GeometricProgressionNthTerm calculates the nth term of a geometric progression
// given the first term (firstTerm), the common ratio (commonRatio), and the position of the term (termPosition).
// It returns the value of the nth term in the geometric progression.
func GeometricProgressionNthTerm(firstTerm, commonRatio float64, termPosition int) float64 {
	return firstTerm * math.Pow(commonRatio, float64(termPosition-1))
}

// GeometricProgressionSum calculates the sum of the first n terms of a geometric progression.
// It takes the first term (firstTerm), the common ratio (commonRatio), and the number of terms (numberOfTerms) as input.
// It returns the sum of the first n terms in the geometric progression.
func GeometricProgressionSum(firstTerm, commonRatio float64, numberOfTerms int) float64 {
	if commonRatio == 1 {
		return firstTerm * float64(numberOfTerms)
	}
	return firstTerm * (1 - math.Pow(commonRatio, float64(numberOfTerms))) / (1 - commonRatio)
}

// FindGeometricProgressionNthTerm calculates the nth term of a geometric progression
// given the first term (firstTerm), the second term (secondTerm), and the position of the term (termPosition).
// It returns the value of the nth term in the geometric progression.
func FindGeometricProgressionNthTerm(firstTerm, secondTerm float64, termPosition int) float64 {
	commonRatio := secondTerm / firstTerm
	return firstTerm * math.Pow(commonRatio, float64(termPosition-1))
}

// FindLengthOfGeometricProgression calculates the number of terms in a geometric progression
// given the first term (firstTerm), the second term (secondTerm), and the last term (lastTerm).
// It returns the number of terms in the geometric progression.
func FindLengthOfGeometricProgression(firstTerm, secondTerm, lastTerm float64) int {
	commonRatio := secondTerm / firstTerm
	termPosition := math.Log(lastTerm/firstTerm)/math.Log(commonRatio) + 1
	return int(termPosition)
}
