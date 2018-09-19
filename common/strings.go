package common

import "strconv"

// NewStringIfNotEmpty takes strings a and b, and returns a unless
// string b is not empty.
func NewStringIfNotEmpty(original string, newString string) string {
	if newString != "" {
		return newString
	}
	return original
}

// NewStringIfNotZero takes string a and int b, and returns a unless
// int b is not zero, in which case it returns Itoa(b).
func NewStringIfNotZero(original string, newString int) string {
	if newString != 0 {
		return strconv.Itoa(newString)
	}
	return original
}
