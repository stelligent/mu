package aws

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codepipeline/codepipelineiface"
	"github.com/stelligent/mu/common"
)

type codePipelineManager struct {
	codePipelineAPI codepipelineiface.CodePipelineAPI
}

func newPipelineManager(sess *session.Session) (common.PipelineManager, error) {
	log.Debug("Connecting to CodePipeline service")
	codePipelineAPI := codepipeline.New(sess)

	return &codePipelineManager{
		codePipelineAPI: codePipelineAPI,
	}, nil
}

// ListState get the state of the pipeline
func (cplMgr *codePipelineManager) ListState(pipelineName string) ([]common.PipelineStageState, error) {
	cplAPI := cplMgr.codePipelineAPI

	params := &codepipeline.GetPipelineStateInput{
		Name: aws.String(pipelineName),
	}

	log.Debugf("Searching for pipeline state for pipeline named '%s'", pipelineName)

	output, err := cplAPI.GetPipelineState(params)
	if err != nil {
		return nil, err
	}

	stageStates := make([]common.PipelineStageState, len(output.StageStates))
	for i, stageState := range output.StageStates {
		stageStates[i] = stageState
	}

	return stageStates, nil
}

// GetPipeline get the config of the pipeline
func (cplMgr *codePipelineManager) GetPipeline(pipelineName string) (*codepipeline.GetPipelineOutput, error) {
	cplAPI := cplMgr.codePipelineAPI

	params := &codepipeline.GetPipelineInput{
		Name: aws.String(pipelineName),
	}

	log.Debugf("Searching for pipeline config for pipeline named '%s'", pipelineName)

	output, err := cplAPI.GetPipeline(params)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (cplMgr *codePipelineManager) GetGitInfo(pipelineName string) (common.GitInfo, error) {
	stageStates, err := cplMgr.ListState(pipelineName)
	if err != nil {
		return common.GitInfo{}, err
	}

	var gitInfo common.GitInfo

	codeCommitRegex := regexp.MustCompile("^http(s?)://.+\\.console\\.aws\\.amazon\\.com/codecommit/home#/repository/([^/]+)/.+$")
	gitHubRegex := regexp.MustCompile("^http(s?)://github\\.com/([^/]+)/([^/]+)/.+$")
	s3Regex := regexp.MustCompile("^http(s?)://console\\.aws\\.amazon\\.com/s3/home\\?#$")

	for _, stageState := range stageStates {
		for _, actionState := range stageState.ActionStates {
			if aws.StringValue(actionState.ActionName) == "Source" {
				entityURL := common.StringValue(actionState.EntityUrl)

				if matches := codeCommitRegex.FindStringSubmatch(entityURL); matches != nil {
					gitInfo.Provider = "CodeCommit"
					gitInfo.RepoName = matches[2]
					gitInfo.Slug = gitInfo.RepoName
				} else if matches := gitHubRegex.FindStringSubmatch(entityURL); matches != nil {
					gitInfo.Provider = "GitHub"
					gitInfo.RepoName = matches[3]
					gitInfo.Slug = fmt.Sprintf("%s/%s", matches[2], matches[3])
				} else if matches := s3Regex.FindStringSubmatch(entityURL); matches != nil {
					pipeline, err := cplMgr.GetPipeline(pipelineName)
					if err != nil {
						return common.GitInfo{}, err
					}
					gitInfo.Provider = "S3"
					gitInfo.RepoName = fmt.Sprintf("%v/%v", *pipeline.Pipeline.Stages[0].Actions[0].Configuration["S3Bucket"], *pipeline.Pipeline.Stages[0].Actions[0].Configuration["S3ObjectKey"])
					gitInfo.Slug = gitInfo.RepoName
				} else {
					return gitInfo, fmt.Errorf("Unable to parse entity url: %s", entityURL)
				}

				if actionState.CurrentRevision != nil && actionState.CurrentRevision.RevisionId != nil {
					// Remove invalid characters from RevisionID
					replacer := strings.NewReplacer(".", "", "_", "", "-", "")
					*actionState.CurrentRevision.RevisionId = replacer.Replace(*actionState.CurrentRevision.RevisionId)
					gitInfo.Revision = aws.StringValue(actionState.CurrentRevision.RevisionId)
				}
				return gitInfo, nil
			}
		}
	}

	return gitInfo, fmt.Errorf("Can not obtain git information from CodePipeline: %s", pipelineName)
}
