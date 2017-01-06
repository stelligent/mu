package common

import (
	"testing"
	"github.com/stretchr/testify/assert"

)

func TestLoadConfig(t *testing.T) {
	assert := assert.New(t)

	config := LoadConfig()

	assert.NotNil(config)
}