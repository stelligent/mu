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

// NewMapElementIfNotEmpty adds a new element to map given by destMap if the
// value given in newString is not empty
func NewMapElementIfNotEmpty(destMap map[string]string, destElement string, newString string) {
	if newString != "" {
		destMap[destElement] = newString
	}
}

// NewMapElementIfNotZero adds a new element to map given by destMap if the
// value given in newInt is not empty (newInt is converted to string)
func NewMapElementIfNotZero(destMap map[string]string, destElement string, newInt int) {
	if newInt != 0 {
		destMap[destElement] = strconv.Itoa(newInt)
	}
}

// NewStringIfNotZero takes string a and int b, and returns a unless
// int b is not zero, in which case it returns Itoa(b).
func NewStringIfNotZero(original string, newString int) string {
	if newString != 0 {
		return strconv.Itoa(newString)
	}
	return original
}
