package utils

import (
	"log"
	"strconv"
)

func ConvStrToFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func ConvFloat64ToStr(f float64, prec int) string {
	return strconv.FormatFloat(f, 'f', prec, 64)
}
