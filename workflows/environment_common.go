package workflows

import (
	"strings"

	"github.com/fatih/color"
	"github.com/stelligent/mu/common"
)

type environmentWorkflow struct {
	environment               *common.Environment
	codeRevision              string
	repoName                  string
	cloudFormationRoleArn     string
	ec2RoleArn                string
	kubernetesResourceManager common.KubernetesResourceManager
	rbacUsers                 []*subjectRoleBinding
	rbacServices              []*subjectRoleBinding
}

type subjectRoleBinding struct {
	Name string
	Role string
}

func colorizeStackStatus(stackStatus string) string {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	var color func(a ...interface{}) string
	if strings.HasSuffix(stackStatus, "_FAILED") {
		color = red
	} else if strings.HasSuffix(stackStatus, "_COMPLETE") {
		color = green
	} else {
		color = blue
	}
	return color(stackStatus)
}

func (workflow *environmentWorkflow) isEcsProvider() Conditional {
	return func() bool {
		return strings.EqualFold(string(workflow.environment.Provider), string(common.EnvProviderEcs))
	}
}

func (workflow *environmentWorkflow) isKubernetesProvider() Conditional {
	return func() bool {
		return strings.EqualFold(string(workflow.environment.Provider), string(common.EnvProviderEks)) ||
			strings.EqualFold(string(workflow.environment.Provider), string(common.EnvProviderEksFargate))
	}
}

func (workflow *environmentWorkflow) isEc2Provider() Conditional {
	return func() bool {
		return strings.EqualFold(string(workflow.environment.Provider), string(common.EnvProviderEc2))
	}
}

func (workflow *environmentWorkflow) connectKubernetes(muNamespace string, provider common.KubernetesResourceManagerProvider) Executor {
	return func() error {
		clusterName := common.CreateStackName(muNamespace, common.StackTypeEnv, workflow.environment.Name)
		kubernetesResourceManager, err := provider.GetResourceManager(clusterName)
		workflow.kubernetesResourceManager = kubernetesResourceManager
		return err
	}
}
