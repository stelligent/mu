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

func TestCodePipelineManager_GetGetInfo(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedCPL)
	m.On("GetPipelineState").Return(
		&codepipeline.GetPipelineStateOutput{
			StageStates: []*codepipeline.StageState{
				{
					ActionStates: []*codepipeline.ActionState{
						{
							ActionName:  aws.String("Source"),
							RevisionUrl: aws.String("https://github.com/dmurawsky/aftp-mu/commit/4e934a1e51476d88d715f421ecd86d93dad02c5b"),
							EntityUrl:   aws.String("https://github.com/dmurawsky/aftp-mu/tree/master"),
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

	gitInfo, err := pipelineManager.GetGitInfo("foo")
	assert.Nil(err)
	assert.Equal("4e934a1e51476d88d715f421ecd86d93dad02c5b", gitInfo.revision)
	assert.Equal("GitHub", gitInfo.provider)
	assert.Equal("aftp-mu", gitInfo.repoName)
	assert.Equal("dmurawsky/aftp-mu", gitInfo.slug)
}

func TestCodePipelineManager_GetGetInfo_CodeCommit(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedCPL)
	m.On("GetPipelineState").Return(
		&codepipeline.GetPipelineStateOutput{
			StageStates: []*codepipeline.StageState{
				{
					ActionStates: []*codepipeline.ActionState{
						{
							ActionName:  aws.String("Source"),
							RevisionUrl: aws.String("https://us-west-2.console.aws.amazon.com/codecommit/home#/repository/banana-service/commit/94cfe74c513178da53e3e33e677db73d8fb5644a"),
							EntityUrl:   aws.String("https://us-west-2.console.aws.amazon.com/codecommit/home#/repository/banana-service/browse/master/--/"),
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

	gitInfo, err := pipelineManager.GetGitInfo("foo")
	assert.Nil(err)
	assert.Equal("CodeCommit", gitInfo.provider)
	assert.Equal("4e934a1e51476d88d715f421ecd86d93dad02c5b", gitInfo.revision)
	assert.Equal("banana-service", gitInfo.repoName)
	assert.Equal("banana-service", gitInfo.slug)
}
