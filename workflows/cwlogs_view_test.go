package workflows

import (
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"testing"
	"time"
)

type mockedLogsManager struct {
	mock.Mock
}

func (m *mockedLogsManager) ViewLogs(logGroup string, searchDuration time.Duration, follow bool, filter string, callback func(string, string, int64)) error {
	args := m.Called(logGroup)
	return args.Error(0)
}

func TestNewEnvironmentLogViewer(t *testing.T) {
	assert := assert.New(t)

	ctx := new(common.Context)
	ctx.Config.Service.Name = "my-service"

	logsManager := new(mockedLogsManager)
	logsManager.On("ViewLogs", "mu-cluster-my-env").Return(nil)
	ctx.LogsManager = logsManager

	searchDuration := 5 * time.Minute
	workflow := NewEnvironmentLogViewer(ctx, searchDuration, false, "my-env", ioutil.Discard, "")

	assert.NotNil(workflow)

	err := workflow()
	assert.Nil(err)

	logsManager.AssertExpectations(t)
	logsManager.AssertNumberOfCalls(t, "ViewLogs", 1)
}

func TestNewServiceLogViewer(t *testing.T) {
	assert := assert.New(t)

	ctx := new(common.Context)

	logsManager := new(mockedLogsManager)
	logsManager.On("ViewLogs", "mu-service-my-service-my-env").Return(nil)
	ctx.LogsManager = logsManager

	searchDuration := 5 * time.Minute
	workflow := NewServiceLogViewer(ctx, searchDuration, false, "my-env", "my-service", ioutil.Discard, "")

	assert.NotNil(workflow)

	err := workflow()
	assert.Nil(err)

	logsManager.AssertExpectations(t)
	logsManager.AssertNumberOfCalls(t, "ViewLogs", 1)
}

func TestNewPipelineLogViewer(t *testing.T) {
	assert := assert.New(t)

	ctx := new(common.Context)

	logsManager := new(mockedLogsManager)
	logsManager.On("ViewLogs", "/aws/codebuild/mu-pipeline-my-service-artifact").Return(nil)
	logsManager.On("ViewLogs", "/aws/codebuild/mu-pipeline-my-service-image").Return(nil)
	logsManager.On("ViewLogs", "/aws/codebuild/mu-pipeline-my-service-deploy-production").Return(nil)
	logsManager.On("ViewLogs", "/aws/codebuild/mu-pipeline-my-service-test-acceptance").Return(nil)
	logsManager.On("ViewLogs", "/aws/codebuild/mu-pipeline-my-service-deploy-acceptance").Return(nil)
	logsManager.On("ViewLogs", "/aws/codebuild/mu-pipeline-my-service-test-production").Return(nil)
	ctx.LogsManager = logsManager

	searchDuration := 5 * time.Minute
	workflow := NewPipelineLogViewer(ctx, searchDuration, false, "my-service", ioutil.Discard, "")

	assert.NotNil(workflow)

	err := workflow()
	assert.Nil(err)

	logsManager.AssertExpectations(t)
	logsManager.AssertNumberOfCalls(t, "ViewLogs", 6)
}
