package main

import (
	"github.com/stelligent/mu/cli"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMainMethod(t *testing.T) {
	var badArgs []string
	os.Args = badArgs
	assert.Panics(t, main)
	os.Args = []string{cli.EnvCmd}
	assert.NotPanics(t, main)
}
