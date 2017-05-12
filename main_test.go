package main

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMainMethod(t *testing.T) {
	var badArgs []string
	os.Args = badArgs
	assert.Panics(t, main)
	os.Args = []string{common.EnvCmd}
	assert.NotPanics(t, main)
}
