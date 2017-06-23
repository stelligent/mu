package common

import (
	"github.com/aws/aws-sdk-go/service/codepipeline"
)

// PipelineStageState a representation of the state of a stage in the pipeline
type PipelineStageState *codepipeline.StageState

// PipelineStateLister for getting cluster instances
type PipelineStateLister interface {
	ListState(pipelineName string) ([]PipelineStageState, error)
}

// PipelineGitInfoGetter for getting the git revision
type PipelineGitInfoGetter interface {
	GetGitInfo(pipelineName string) (GitInfo, error)
}

// GitInfo represents pertinent git information
type GitInfo struct {
	Provider string
	Revision string
	RepoName string
	Slug     string
}

// PipelineManager composite of all cluster capabilities
type PipelineManager interface {
	PipelineStateLister
	PipelineGitInfoGetter
}
