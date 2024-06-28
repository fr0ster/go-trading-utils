package progressions

type (
	NthTermType                     func(firstTerm, commonRatio float64, termPosition int) float64
	DeltaType                       func(firstTerm, secondTerm float64) float64
	SumType                         func(firstTerm, commonRatio float64, numberOfTerms int) float64
	FindNthTermType                 func(firstTerm, secondTerm float64, termPosition int) float64
	FindLengthOfProgressionType     func(firstTerm, secondTerm, lastTerm float64) int
	FindCubicProgressionTthTermType func(firstTerm, lastTerm float64, length int, sum float64, T int) float64
	Progression                     interface {
		NthTerm(firstTerm, commonRatio float64, termPosition int) float64
		Sum(firstTerm, commonRatio float64, numberOfTerms int) float64
		FindNthTerm(firstTerm, secondTerm float64, termPosition int) float64
		FindLengthOfProgression(firstTerm, secondTerm, lastTerm float64) int
	}
)
