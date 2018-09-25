package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabase_GetDatabaseConfig(t *testing.T) {
	assert := assert.New(t)

	db := &Database{
		DatabaseConfig: DatabaseConfig{
			Name:   "name-foo",
			Engine: "postgres-foo",
		},
		EnvironmentConfig: map[string]DatabaseConfig{
			"acceptance": DatabaseConfig{
				Engine:     "mysql-bar",
				EngineMode: "serverless",
			},
		},
	}

	acptConfig := db.GetDatabaseConfig("acceptance")
	prodConfig := db.GetDatabaseConfig("production")

	assert.Equal("name-foo", acptConfig.Name)
	assert.Equal("name-foo", prodConfig.Name)

	assert.Equal("mysql-bar", acptConfig.Engine)
	assert.Equal("postgres-foo", prodConfig.Engine)

	assert.Equal("serverless", acptConfig.EngineMode)
	assert.Equal("", prodConfig.EngineMode)
}
