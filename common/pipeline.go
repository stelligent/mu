package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codepipeline/codepipelineiface"
)

// PipelineStateLister for getting cluster instances
type PipelineStateLister interface {
	ListState(pipelineName string) ([]*codepipeline.StageState, error)
}

// PipelineManager composite of all cluster capabilities
type PipelineManager interface {
	PipelineStateLister
}

type codePipelineManager struct {
	codePipelineAPI codepipelineiface.CodePipelineAPI
}

func newPipelineManager(region string) (PipelineManager, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	log.Debugf("Connecting to CodePipeline service in region:%s", region)
	codePipelineAPI := codepipeline.New(sess, &aws.Config{Region: aws.String(region)})

	return &codePipelineManager{
		codePipelineAPI: codePipelineAPI,
	}, nil
}

// ListState get the state of the pipeline
func (cplMgr *codePipelineManager) ListState(pipelineName string) ([]*codepipeline.StageState, error) {
	cplAPI := cplMgr.codePipelineAPI

	params := &codepipeline.GetPipelineStateInput{
		Name: aws.String(pipelineName),
	}

	log.Debugf("Searching for pipeline state for pipeline named '%s'", pipelineName)

	output, err := cplAPI.GetPipelineState(params)
	if err != nil {
		return nil, err
	}

	return output.StageStates, nil
}
