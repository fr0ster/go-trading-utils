package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

type APIError struct {
	Code int
	Msg  string
}

func ParseAPIError(errMsg error) (*APIError, error) {
	re := regexp.MustCompile(`<APIError> code=(-?\d+), msg=(.*)`)
	matches := re.FindStringSubmatch(errMsg.Error())
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid error message")
	}
	code, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, err
	}
	return &APIError{
		Code: code,
		Msg:  matches[2],
	}, nil
}
