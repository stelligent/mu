package common

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"github.com/aws/aws-sdk-go/service/codepipeline/codepipelineiface"
	"github.com/aws/aws-sdk-go/service/codepipeline"
)

type mockedCPL struct {
	mock.Mock
	codepipelineiface.CodePipelineAPI
}

func (m *mockedCPL) GetPipelineState(input *codepipeline.GetPipelineStateInput) (*codepipeline.GetPipelineStateOutput, error) {
	args := m.Called()
	return args.Get(0).(*codepipeline.GetPipelineStateOutput), args.Error(1)
}

func TestCodePipelineManager_ListState(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedCPL)
	m.On("GetPipelineState").Return(
		&codepipeline.GetPipelineStateOutput{
			StageStates: []*codepipeline.StageState{},
		}, nil)

	pipelineManager := codePipelineManager{
		codePipelineAPI: m,
	}

	states, err := pipelineManager.ListState("foo")
	assert.Nil(err)
	assert.NotNil(states)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "GetPipelineState", 1)
}
