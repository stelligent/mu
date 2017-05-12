package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codepipeline/codepipelineiface"
	"regexp"
)

// PipelineStateLister for getting cluster instances
type PipelineStateLister interface {
	ListState(pipelineName string) ([]*codepipeline.StageState, error)
}

// PipelineGitInfoGetter for getting the git revision
type PipelineGitInfoGetter interface {
	GetGitInfo(pipelineName string) (GitInfo, error)
}

// GitInfo represents pertinent git information
type GitInfo struct {
	provider string
	revision string
	repoName string
	slug     string
}

// PipelineManager composite of all cluster capabilities
type PipelineManager interface {
	PipelineStateLister
	PipelineGitInfoGetter
}

type codePipelineManager struct {
	codePipelineAPI codepipelineiface.CodePipelineAPI
}

func newPipelineManager(sess *session.Session) (PipelineManager, error) {
	log.Debug("Connecting to CodePipeline service")
	codePipelineAPI := codepipeline.New(sess)

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

func (cplMgr *codePipelineManager) GetGitInfo(pipelineName string) (GitInfo, error) {
	stageStates, err := cplMgr.ListState(pipelineName)
	if err != nil {
		return GitInfo{}, err
	}

	var gitInfo GitInfo

	codeCommitRegex := regexp.MustCompile("^http(s?)://.+\\.console\\.aws\\.amazon\\.com/codecommit/home#/repository/([^/]+)/.+$")
	gitHubRegex := regexp.MustCompile("^http(s?)://github\\.com/([^/]+)/([^/]+)/.+$")

	for _, stageState := range stageStates {
		for _, actionState := range stageState.ActionStates {
			if aws.StringValue(actionState.ActionName) == "Source" {
				entityURL := aws.StringValue(actionState.EntityUrl)

				if matches := codeCommitRegex.FindStringSubmatch(entityURL); matches != nil {
					gitInfo.provider = "CodeCommit"
					gitInfo.repoName = matches[2]
					gitInfo.slug = gitInfo.repoName
				} else if matches := gitHubRegex.FindStringSubmatch(entityURL); matches != nil {
					gitInfo.provider = "GitHub"
					gitInfo.repoName = matches[3]
					gitInfo.slug = fmt.Sprintf("%s/%s", matches[2], matches[3])
				} else {
					return gitInfo, fmt.Errorf("Unable to parse entity url: %s", entityURL)
				}

				gitInfo.revision = aws.StringValue(actionState.CurrentRevision.RevisionId)
				return gitInfo, nil
			}
		}
	}

	return gitInfo, fmt.Errorf("Can not obtain git information from CodePipeline: %s", pipelineName)
}
