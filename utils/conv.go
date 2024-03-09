package utils

import "strconv"

func ConvStrToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func ConvFloat64ToStr(f float64, prec int) string {
	return strconv.FormatFloat(f, 'f', prec, 64)
}
