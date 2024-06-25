package utils

import "math"

// ArithmeticProgressionNthTerm calculates the nth term of an arithmetic progression
// given the first term (firstTerm), the common difference (commonDifference), and the position of the term (termPosition).
// It returns the value of the nth term in the arithmetic progression.
func ArithmeticProgressionNthTerm(firstTerm, commonDifference float64, termPosition int) float64 {
	return firstTerm + float64(termPosition-1)*commonDifference
}

// FindArithmeticProgressionNthTerm calculates the nth term of an arithmetic progression
// given the first term (firstTerm), the second term (secondTerm), and the position of the term (termPosition).
// It returns the value of the nth term in the arithmetic progression.
func FindArithmeticProgressionNthTerm(firstTerm, secondTerm float64, termPosition int) float64 {
	commonDifference := secondTerm - firstTerm
	return firstTerm + float64(termPosition-1)*commonDifference
}

// ArithmeticProgressionSum calculates the sum of the first n terms of an arithmetic progression.
// It takes the first term (firstTerm), the common difference (commonDifference), and the number of terms (numberOfTerms) as input.
// It returns the sum of the first n terms in the arithmetic progression.
func ArithmeticProgressionSum(firstTerm, commonDifference float64, numberOfTerms int) float64 {
	return float64(numberOfTerms) / 2 * (2*firstTerm + float64(numberOfTerms-1)*commonDifference)
}

// FindLengthOfArithmeticProgression calculates the number of terms in an arithmetic progression
// given the first term (firstTerm), the second term (secondTerm), and the last term (lastTerm).
// It returns the number of terms in the arithmetic progression.
func FindLengthOfArithmeticProgression(firstTerm, secondTerm, lastTerm float64) int {
	commonDifference := secondTerm - firstTerm
	numberOfTerms := (lastTerm-firstTerm)/commonDifference + 1
	return int(numberOfTerms)
}

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
