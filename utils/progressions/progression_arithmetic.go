package progressions

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

// FindArithmeticProgressionTthTerm calculates the Tth term of an arithmetic progression
// given the first term (firstTerm), the last term (lastTerm), the length of the progression (length),
// and the sum of the progression (sum). It returns the value of the Tth term in the arithmetic progression.
func FindArithmeticProgressionTthTerm(firstTerm, lastTerm float64, length int, sum float64, T int) float64 {
	// Calculate the common difference of the arithmetic progression
	commonDifference := (lastTerm - firstTerm) / float64(length-1)
	// Calculate the Tth term using the formula: TthTerm = firstTerm + (T-1)*commonDifference
	// Where TthTerm is the Tth term, firstTerm is the first term, and commonDifference is the common difference.
	TthTerm := firstTerm + float64(T-1)*commonDifference
	return TthTerm
}
