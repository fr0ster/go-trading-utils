package utils_test

import (
	"testing"

	"github.com/fr0ster/go-trading-utils/utils"
)

func TestRoundToDecimalPlace(t *testing.T) {
	tests := []struct {
		num            float64
		decimalPlaces  float64
		expectedResult float64
	}{
		{3.14159, 2, 3.14},
		{2.71828, 3, 2.718},
		{1.23456789, 4, 1.2346},
		{0.123456789, 6, 0.123457},
		{111.0, 0, 111.0},
		{111.0, -1, 110.0},
		{111.0, -2, 100.0},
		{11100.0, -3, 11000.0},
		{11100.0, -4, 10000.0},
		{11100.0, -5, 0.0},
		{999999.0, -5, 1000000.0},
	}

	for _, test := range tests {
		result := utils.RoundToDecimalPlace(test.num, test.decimalPlaces)
		if result != test.expectedResult {
			t.Errorf("Expected %f, but got %f for input (%f, %f)", test.expectedResult, result, test.num, test.decimalPlaces)
		}
	}
}
