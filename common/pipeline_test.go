package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codepipeline/codepipelineiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
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

func TestCodePipelineManager_GetCurrentRevision(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedCPL)
	m.On("GetPipelineState").Return(
		&codepipeline.GetPipelineStateOutput{
			StageStates: []*codepipeline.StageState{
				{
					ActionStates: []*codepipeline.ActionState{
						{
							ActionName: aws.String("Source"),
							RevisionUrl: aws.String("https://github.com/dmurawsky/aftp-mu/commit/4e934a1e51476d88d715f421ecd86d93dad02c5b"),
							EntityUrl: aws.String("https://github.com/dmurawsky/aftp-mu/tree/master"),
							LatestExecution: &codepipeline.ActionExecution{
								Status: aws.String("Succeeded"),
							},
						  CurrentRevision: &codepipeline.ActionRevision{
								RevisionId: aws.String("4e934a1e51476d88d715f421ecd86d93dad02c5b"),
							},
						},
					},
				},
			},
		},
		nil,
	)

	pipelineManager := codePipelineManager{
		codePipelineAPI: m,
	}

	revision, err := pipelineManager.GetCurrentRevision("foo")
	assert.Nil(err)
	assert.Equal("4e934a1e51476d88d715f421ecd86d93dad02c5b", revision)
}
