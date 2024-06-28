package progressions_test

import (
	"testing"

	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"
	"github.com/stretchr/testify/assert"
)

func TestCubicProgressionNthTerm(t *testing.T) {
	// Test case 1: First term is 1, common ratio is 2, term position is 5
	// Expected output: 4096
	result := progressions.CubicProgressionNthTerm(1, 2, 5)
	expected := 4096.0
	assert.Equal(t, expected, result)

	// Test case 2: First term is -3, common ratio is -1, term position is 4
	// Expected output: 3
	result = progressions.CubicProgressionNthTerm(-3, -1, 4)
	expected = 3.0
	assert.Equal(t, expected, result)

	// Test case 3: First term is 0, common ratio is 0.5, term position is 10
	// Expected output: 0
	result = progressions.CubicProgressionNthTerm(0, 0.5, 10)
	expected = 0.0
	assert.Equal(t, expected, result)
}

func TestFindCubicProgressionNthTerm(t *testing.T) {
	// Test case 1: First term is 1, second term is 3, term position is 4
	// Expected output: 27.00000000000003
	result := progressions.FindCubicProgressionNthTerm(1, 3, 4)
	expected := 27.00000000000003
	assert.Equal(t, expected, result)

	// Test case 2: First term is -2, second term is -4, term position is 6
	// Expected output: -63.99999999999984
	result = progressions.FindCubicProgressionNthTerm(-2, -4, 6)
	expected = -63.99999999999984
	assert.Equal(t, expected, result)
}

func TestCubicProgressionSum(t *testing.T) {
	// Test case 1: First term is 1, common ratio is 2, number of terms is 5
	// Expected output: 4681
	result := progressions.CubicProgressionSum(1, 2, 5)
	expected := 4681.0
	assert.Equal(t, expected, result)

	// Test case 2: First term is -3, common ratio is -1, number of terms is 4
	// Expected output: 0
	result = progressions.CubicProgressionSum(-3, -1, 4)
	expected = 0
	assert.Equal(t, expected, result)

	// Test case 3: First term is 0, common ratio is 0.5, number of terms is 10
	// Expected output: 0
	result = progressions.CubicProgressionSum(0, 0.5, 10)
	expected = 0
	assert.Equal(t, expected, result)
}

func TestFindLengthOfCubicProgression(t *testing.T) {
	// Test case 1: First term is 1, second term is 3, last term is 9
	// Expected output: 2
	result := progressions.FindLengthOfCubicProgression(1, 3, 9)
	expected := 2
	assert.Equal(t, expected, result)

	// Test case 2: First term is -2, second term is -4, last term is -14
	// Expected output: 2
	result = progressions.FindLengthOfCubicProgression(-2, -4, -14)
	expected = 2
	assert.Equal(t, expected, result)
}

func TestFindCubicProgressionTthTerm(t *testing.T) {
	// Test case 1: First term is 1, last term is 9, length is 4, sum is 19, T is 3
	// Expected output: 7
	result := progressions.FindCubicProgressionTthTerm(1, 9, 4, 19, 3)
	expected := 531441.0
	assert.Equal(t, expected, result)

	// Test case 2: First term is -2, last term is -14, length is 6, sum is -27, T is 5
	// Expected output: -10
	result = progressions.FindCubicProgressionTthTerm(-2, -14, 6, -27, 5)
	expected = -2.7682574402e+10
	assert.Equal(t, expected, result)

	// Test case 3: First term is 0, last term is 0, length is 1, sum is 0, T is 1
	// Expected output: 0
	result = progressions.FindCubicProgressionTthTerm(0, 0, 1, 0, 1)
	expected = 0.0
	assert.Equal(t, expected, result)
}
