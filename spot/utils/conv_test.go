package utils_test

import (
	"testing"

	"github.com/fr0ster/go-binance-utils/spot/utils"
)

func TestConvStrToFloat64(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"10.5", 10.5},
		{"-5.2", -5.2},
		{"0", 0},
		{"invalid", 0}, // Invalid input should return 0
	}

	for _, test := range tests {
		result := utils.ConvStrToFloat64(test.input)
		if result != test.expected {
			t.Errorf("ConvStrToFloat64(%s) = %f, expected %f", test.input, result, test.expected)
		}
	}
}

func TestConvFloat64ToStrDefault(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{10.5, "10.50000000"},
		{-5.2, "-5.20000000"},
		{0, "0.00000000"},
	}

	for _, test := range tests {
		result := utils.ConvFloat64ToStrDefault(test.input)
		if result != test.expected {
			t.Errorf("ConvFloat64ToStrDefault(%f) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestConvFloat64ToStr(t *testing.T) {
	tests := []struct {
		input    float64
		prec     int
		expected string
	}{
		{10.5, 2, "10.50"},
		{-5.2, 1, "-5.2"},
		{0, 0, "0"},
	}

	for _, test := range tests {
		result := utils.ConvFloat64ToStr(test.input, test.prec)
		if result != test.expected {
			t.Errorf("ConvFloat64ToStr(%f, %d) = %s, expected %s", test.input, test.prec, result, test.expected)
		}
	}
}
