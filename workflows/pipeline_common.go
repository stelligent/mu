package workflows

import "github.com/stelligent/mu/common"

type pipelineWorkflow struct {
	serviceName    string
	pipelineConfig *common.Pipeline
}
