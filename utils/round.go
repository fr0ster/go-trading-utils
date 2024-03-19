package utils

import "math"

func RoundToDecimalPlace(x float64, exp float64) float64 {
	if exp >= 0 {
		pow := math.Pow(10, float64(exp))
		return math.Round(x*pow) / pow
	} else {
		pow := math.Pow(10, float64(-exp))
		return math.Round(x/pow) * pow
	}
}
