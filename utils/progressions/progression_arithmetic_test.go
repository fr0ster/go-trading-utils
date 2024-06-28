package progressions_test

import (
	"testing"

	progressions "github.com/fr0ster/go-trading-utils/utils/progressions"
)

func TestArithmeticProgressionNthTerm(t *testing.T) {
	// Test case 1: First term is 2, common difference is 3, term position is 4
	// Expected output: 11
	result := progressions.ArithmeticProgressionNthTerm(2, 3, 4)
	if result != 11 {
		t.Errorf("Expected 11, but got %f", result)
	}

	// Test case 2: First term is -5, common difference is -2, term position is 6
	// Expected output: -15
	result = progressions.ArithmeticProgressionNthTerm(-5, -2, 6)
	if result != -15 {
		t.Errorf("Expected -15, but got %f", result)
	}

	// Add more test cases here...
}

func TestFindArithmeticProgressionNthTerm(t *testing.T) {
	// Test case 1: First term is 2, second term is 5, term position is 4
	// Expected output: 1
	result := progressions.FindArithmeticProgressionNthTerm(2, 5, 4)
	if result != 11 {
		t.Errorf("Expected 11, but got %f", result)
	}

	// Test case 2: First term is -3, second term is -7, term position is 6
	// Expected output: -23
	result = progressions.FindArithmeticProgressionNthTerm(-3, -7, 6)
	if result != -23 {
		t.Errorf("Expected -23, but got %f", result)
	}

	// Add more test cases here...
}

func TestArithmeticProgressionSum(t *testing.T) {
	// Test case 1: First term is 1, common difference is 2, number of terms is 5
	// Expected output: 25
	result := progressions.ArithmeticProgressionSum(1, 2, 5)
	if result != 25 {
		t.Errorf("Expected 25, but got %f", result)
	}

	// Test case 2: First term is -4, common difference is -3, number of terms is 6
	// Expected output: -69
	result = progressions.ArithmeticProgressionSum(-4, -3, 6)
	if result != -69 {
		t.Errorf("Expected -69, but got %f", result)
	}

	// Add more test cases here...
}

func TestFindLengthOfArithmeticProgression(t *testing.T) {
	// Test case 1: First term is 3, second term is 7, last term is 23
	// Expected output: 6
	result := progressions.FindLengthOfArithmeticProgression(3, 7, 23)
	if result != 6 {
		t.Errorf("Expected 6, but got %d", result)
	}

	// Test case 2: First term is -2, second term is -5, last term is -20
	// Expected output: 7
	result = progressions.FindLengthOfArithmeticProgression(-2, -5, -20)
	if result != 7 {
		t.Errorf("Expected 7, but got %d", result)
	}

	// Add more test cases here...
}

func TestFindArithmeticProgressionTthTerm(t *testing.T) {
	// Test case 1: First term is 2, last term is 20, length is 10, sum is 110, T is 5
	// Expected output: 10
	result := progressions.FindArithmeticProgressionTthTerm(2, 20, 10, 110, 5)
	if result != 10 {
		t.Errorf("Expected 10, but got %f", result)
	}

	// Test case 2: First term is -3, last term is -30, length is 10, sum is -165, T is 8
	// Expected output: -24
	result = progressions.FindArithmeticProgressionTthTerm(-3, -30, 10, -165, 8)
	if result != -24 {
		t.Errorf("Expected -24, but got %f", result)
	}

	// Add more test cases here...
}
