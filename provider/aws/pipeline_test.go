package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codepipeline/codepipelineiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedCPL struct {
	mock.Mock
	codepipelineiface.CodePipelineAPI
}

func (m *mockedCPL) GetPipelineState(input *codepipeline.GetPipelineStateInput) (*codepipeline.GetPipelineStateOutput, error) {
	args := m.Called()
	return args.Get(0).(*codepipeline.GetPipelineStateOutput), args.Error(1)
}

func (m *mockedCPL) GetPipeline(input *codepipeline.GetPipelineInput) (*codepipeline.GetPipelineOutput, error) {
	args := m.Called()
	return args.Get(0).(*codepipeline.GetPipelineOutput), args.Error(1)
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

func TestCodePipelineManager_GetPipeline(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedCPL)

	m.On("GetPipeline").Return(
		&codepipeline.GetPipelineOutput{
			Pipeline: &codepipeline.PipelineDeclaration{
				Stages: []*codepipeline.StageDeclaration{
					{
						Actions: []*codepipeline.ActionDeclaration{
							{
								Configuration: map[string]*string{
									"S3Bucket":    aws.String("mu-test-bucket"),
									"S3ObjectKey": aws.String("artifacts/latest.zip"),
								},
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

	output, err := pipelineManager.GetPipeline("foo")
	assert.Nil(err)
	assert.NotNil(output)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "GetPipeline", 1)
}

func TestCodePipelineManager_GetGitInfo(t *testing.T) {
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
	assert.Equal("4e934a1e51476d88d715f421ecd86d93dad02c5b", gitInfo.Revision)
	assert.Equal("GitHub", gitInfo.Provider)
	assert.Equal("aftp-mu", gitInfo.RepoName)
	assert.Equal("dmurawsky/aftp-mu", gitInfo.Slug)
}

func TestCodePipelineManager_GetGitInfo_CodeCommit(t *testing.T) {
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
	assert.Equal("CodeCommit", gitInfo.Provider)
	assert.Equal("4e934a1e51476d88d715f421ecd86d93dad02c5b", gitInfo.Revision)
	assert.Equal("banana-service", gitInfo.RepoName)
	assert.Equal("banana-service", gitInfo.Slug)
}

func TestCodePipelineManager_GetGitInfo_S3(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedCPL)
	m.On("GetPipelineState").Return(
		&codepipeline.GetPipelineStateOutput{
			StageStates: []*codepipeline.StageState{
				{
					ActionStates: []*codepipeline.ActionState{
						{
							ActionName: aws.String("Source"),
							EntityUrl:  aws.String("https://console.aws.amazon.com/s3/home?#"),
							LatestExecution: &codepipeline.ActionExecution{
								Status: aws.String("Succeeded"),
							},
							CurrentRevision: &codepipeline.ActionRevision{
								RevisionId: aws.String(".N-N_2QtU0xO2HlXFWnMb1C8bzsoP9eiG7q"),
							},
						},
					},
				},
			},
		},
		nil,
	)

	m.On("GetPipeline").Return(
		&codepipeline.GetPipelineOutput{
			Pipeline: &codepipeline.PipelineDeclaration{
				Stages: []*codepipeline.StageDeclaration{
					{
						Actions: []*codepipeline.ActionDeclaration{
							{
								Configuration: map[string]*string{
									"S3Bucket":    aws.String("mu-test-bucket"),
									"S3ObjectKey": aws.String("artifacts/latest.zip"),
								},
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
	assert.Equal("S3", gitInfo.Provider)
	assert.Equal("NN2QtU0xO2HlXFWnMb1C8bzsoP9eiG7q", gitInfo.Revision)
	assert.Equal("mu-test-bucket/artifacts/latest.zip", gitInfo.RepoName)
	assert.Equal("mu-test-bucket/artifacts/latest.zip", gitInfo.Slug)
}
