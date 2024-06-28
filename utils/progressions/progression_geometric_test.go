package progressions_test

import (
	"testing"

	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"
	"github.com/stretchr/testify/assert"
)

func TestGeometricProgressionNthTerm(t *testing.T) {
	// Test case 1: common ratio is 2, first term is 1, term position is 5
	expectedResult := 16.0
	result := progressions.GeometricProgressionNthTerm(1, 2, 5)
	assert.Equal(t, expectedResult, result)

	// Test case 2: common ratio is 0.5, first term is 10, term position is 3
	expectedResult = 2.5
	result = progressions.GeometricProgressionNthTerm(10, 0.5, 3)
	assert.Equal(t, expectedResult, result)

	// Add more test cases here...
}

func TestGeometricProgressionSum(t *testing.T) {
	// Test case 1: common ratio is 3, first term is 1, number of terms is 4
	expectedResult := 40.0
	result := progressions.GeometricProgressionSum(1, 3, 4)
	assert.Equal(t, expectedResult, result)

	// Test case 2: common ratio is 0.5, first term is 10, number of terms is 5
	expectedResult = 19.375
	result = progressions.GeometricProgressionSum(10, 0.5, 5)
	assert.Equal(t, expectedResult, result)

	// Add more test cases here...
}

func TestFindGeometricProgressionNthTerm(t *testing.T) {
	// Test case 1: first term is 2, second term is 4, term position is 5
	expectedResult := 32.0
	result := progressions.FindGeometricProgressionNthTerm(2, 4, 5)
	assert.Equal(t, expectedResult, result)

	// Test case 2: first term is 5, second term is 10, term position is 3
	expectedResult = 20.0
	result = progressions.FindGeometricProgressionNthTerm(5, 10, 3)
	assert.Equal(t, expectedResult, result)

	// Add more test cases here...
}

func TestFindLengthOfGeometricProgression(t *testing.T) {
	// Test case 1: first term is 2, second term is 4, last term is 64
	expectedResult := 6
	result := progressions.FindLengthOfGeometricProgression(2, 4, 64)
	assert.Equal(t, expectedResult, result)

	// Test case 2: first term is 5, second term is 10, last term is 320
	expectedResult = 7
	result = progressions.FindLengthOfGeometricProgression(5, 10, 320)
	assert.Equal(t, expectedResult, result)

	// Add more test cases here...
}

func TestFindGeometricProgressionTthTerm(t *testing.T) {
	// Test case 1: first term is 2, last term is 64, length is 7, sum is 126, T is 3
	expectedResult := 6.3496042078727974
	result := progressions.FindGeometricProgressionTthTerm(2, 64, 7, 126, 3)
	assert.Equal(t, expectedResult, result)

	// Test case 2: first term is 5, last term is 320, length is 6, sum is 630, T is 4
	expectedResult = 60.62866266041594
	result = progressions.FindGeometricProgressionTthTerm(5, 320, 6, 630, 4)
	assert.Equal(t, expectedResult, result)

	// Add more test cases here...
}
