package utils_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/utils"
)

func TestRoundToDecimalPlace(t *testing.T) {
	tests := []struct {
		num            float64
		decimalPlaces  int
		expectedResult float64
	}{
		{3.14159, 2, 3.14},
		{2.71828, 3, 2.718},
		{1.23456789, 4, 1.2346},
		{0.123456789, 6, 0.123457},
	}

	for _, test := range tests {
		result := utils.RoundToDecimalPlace(test.num, test.decimalPlaces)
		if result != test.expectedResult {
			t.Errorf("Expected %f, but got %f for input (%f, %d)", test.expectedResult, result, test.num, test.decimalPlaces)
		}
	}
}
