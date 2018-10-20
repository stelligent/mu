package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainMethod(t *testing.T) {
	var badArgs []string
	os.Args = badArgs
	assert.Panics(t, main)
	os.Args = []string{"mu", "-v"}
	assert.NotPanics(t, main)
}
