package utils

import (
	"fmt"
	"strconv"
)

func ConvStrToFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func ConvFloat64ToStrDefault(f float64) string {
	return strconv.FormatFloat(f, 'f', 8, 64)
}

func ConvFloat64ToStr(f float64, prec int) string {
	return strconv.FormatFloat(f, 'f', prec, 64)
}

func ConvFloat64ToStrNoExtraZeros(f float64) string {
	return fmt.Sprintf("%g", f)
}
