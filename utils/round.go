package utils

import "math"

func RoundToDecimalPlace(num float64, decimalPlaces int) float64 {
	if decimalPlaces < 0 {
		return num
	} else {
		multiplier := math.Pow(10, float64(decimalPlaces))
		return math.Round(num*multiplier) / multiplier
	}
}
