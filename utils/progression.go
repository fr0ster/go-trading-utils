package utils

type Progression interface {
	NthTerm(firstTerm, commonRatio float64, termPosition int) float64
	Sum(firstTerm, commonRatio float64, numberOfTerms int) float64
	FindNthTerm(firstTerm, secondTerm float64, termPosition int) float64
	FindLengthOfProgression(firstTerm, secondTerm, lastTerm float64) int
}
