package progressions_test

import (
	"testing"

	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"
	"github.com/stretchr/testify/assert"
)

func TestCubicRootProgressionNthTerm(t *testing.T) {
	// Test case 1: commonRatio = 2, termPosition = 3
	firstTerm := 1.0
	commonRatio := 2.0
	termPosition := 3
	expectedResult := 1.5874010519681994
	result := progressions.CubicRootProgressionNthTerm(firstTerm, commonRatio, termPosition)
	assert.Equal(t, expectedResult, result)

	// Test case 2: commonRatio = 0.5, termPosition = 5
	firstTerm = 2.0
	commonRatio = 0.5
	termPosition = 5
	expectedResult = 0.7937005259840998
	result = progressions.CubicRootProgressionNthTerm(firstTerm, commonRatio, termPosition)
	assert.Equal(t, expectedResult, result)

	// Test case 3: commonRatio = -1, termPosition = 4
	firstTerm = 3.0
	commonRatio = -1.0
	termPosition = 4
	expectedResult = -3.0
	result = progressions.CubicRootProgressionNthTerm(firstTerm, commonRatio, termPosition)
	assert.Equal(t, expectedResult, result)
}

func TestFindCubicRootProgressionNthTerm(t *testing.T) {
	// Test case 1: firstTerm = 1, secondTerm = 8, termPosition = 4
	firstTerm := 1.0
	secondTerm := 8.0
	termPosition := 4
	expectedResult := 2.0
	result := progressions.FindCubicRootProgressionNthTerm(firstTerm, secondTerm, termPosition)
	assert.Equal(t, expectedResult, result)

	// Test case 2: firstTerm = 2, secondTerm = 0.125, termPosition = 5
	firstTerm = 2.0
	secondTerm = 0.125
	termPosition = 5
	expectedResult = 0.583265
	result = progressions.FindCubicRootProgressionNthTerm(firstTerm, secondTerm, termPosition)
	assert.Equal(t, expectedResult, result)

	// Test case 3: firstTerm = -3, secondTerm = 3, termPosition = 4
	firstTerm = -3.0
	secondTerm = 3.0
	termPosition = 4
	expectedResult = 3.000000
	result = progressions.FindCubicRootProgressionNthTerm(firstTerm, secondTerm, termPosition)
	assert.Equal(t, expectedResult, result)
}

func TestCubicRootProgressionSum(t *testing.T) {
	// Test case 1: commonRatio = 2, numberOfTerms = 3
	firstTerm := 1.0
	commonRatio := 2.0
	numberOfTerms := 3
	expectedResult := 3.8473221018630723
	result := progressions.CubicRootProgressionSum(firstTerm, commonRatio, numberOfTerms)
	assert.Equal(t, expectedResult, result)

	// Test case 2: commonRatio = 0.5, numberOfTerms = 5
	firstTerm = 2.0
	commonRatio = 0.5
	numberOfTerms = 5
	expectedResult = 6.641022627847173
	result = progressions.CubicRootProgressionSum(firstTerm, commonRatio, numberOfTerms)
	assert.Equal(t, expectedResult, result)

	// Test case 3: commonRatio = -1, numberOfTerms = 4
	firstTerm = 3.0
	commonRatio = 1.0
	numberOfTerms = 4
	expectedResult = 12
	result = progressions.CubicRootProgressionSum(firstTerm, commonRatio, numberOfTerms)
	assert.Equal(t, expectedResult, result)
}

func TestFindLengthOfCubicRootProgression(t *testing.T) {
	// Test case 1: firstTerm = 1, secondTerm = 8, lastTerm = 64
	firstTerm := 1.0
	secondTerm := 8.0
	lastTerm := 64.0
	expectedResult := 18
	result := progressions.FindLengthOfCubicRootProgression(firstTerm, secondTerm, lastTerm)
	assert.Equal(t, expectedResult, result)

	// Test case 2: firstTerm = 2, secondTerm = 0.125, lastTerm = 0.008
	firstTerm = 2.0
	secondTerm = 0.125
	lastTerm = 0.008
	expectedResult = 18
	result = progressions.FindLengthOfCubicRootProgression(firstTerm, secondTerm, lastTerm)
	assert.Equal(t, expectedResult, result)

	// Test case 3: firstTerm = 3, secondTerm = 9, lastTerm = 26
	firstTerm = 3.0
	secondTerm = 9.0
	lastTerm = 64.0
	expectedResult = 26
	result = progressions.FindLengthOfCubicRootProgression(firstTerm, secondTerm, lastTerm)
	assert.Equal(t, expectedResult, result)
}

func TestFindCubicRootProgressionTthTerm(t *testing.T) {
	// Test case 1: firstTerm = 1, lastTerm = 64, length = 4, sum = 85, T = 3
	firstTerm := 1.0
	lastTerm := 64.0
	length := 4
	sum := 85.0
	T := 3
	expectedResult := 15.999999999999993
	result := progressions.FindCubicRootProgressionTthTerm(firstTerm, lastTerm, length, sum, T)
	assert.Equal(t, expectedResult, result)

	// Test case 2: firstTerm = 2, lastTerm = 0.008, length = 5, sum = 2.135, T = 4
	firstTerm = 2.0
	lastTerm = 0.008
	length = 5
	sum = 2.135
	T = 4
	expectedResult = 0.03181082915068204
	result = progressions.FindCubicRootProgressionTthTerm(firstTerm, lastTerm, length, sum, T)
	assert.Equal(t, expectedResult, result)

	// Test case 3: firstTerm = -3, lastTerm = -3, length = 1, sum = -3, T = 1
	firstTerm = -3.0
	lastTerm = -3.0
	length = 1
	sum = -3.0
	T = 1
	expectedResult = -3.0
	result = progressions.FindCubicRootProgressionTthTerm(firstTerm, lastTerm, length, sum, T)
	assert.Equal(t, expectedResult, result)
}
