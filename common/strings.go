package common

import "strconv"

func NewStringIfNotEmpty(original string, newString string) string {
	if newString != "" {
		return newString
	}
	return original
}

func NewStringIfNotZero(original string, newString int) string {
	if newString != 0 {
		return strconv.Itoa(newString)
	}
	return original
}
