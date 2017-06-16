package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
)

// ElbRule for the rules in ELB listeners
type ElbRule *elbv2.Rule

// ElbRuleLister for getting cluster instances
type ElbRuleLister interface {
	ListRules(listenerArn string) ([]ElbRule, error)
}

// ElbManager composite of all cluster capabilities
type ElbManager interface {
	ElbRuleLister
}

type elbv2Manager struct {
	elbAPI elbv2iface.ELBV2API
}

func newElbv2Manager(sess *session.Session) (ElbManager, error) {
	log.Debug("Connecting to ELBv2 service")
	elbAPI := elbv2.New(sess)

	return &elbv2Manager{
		elbAPI: elbAPI,
	}, nil
}

// ListState get the state of the pipeline
func (elbMgr *elbv2Manager) ListRules(listenerArn string) ([]ElbRule, error) {
	elbAPI := elbMgr.elbAPI

	params := &elbv2.DescribeRulesInput{
		ListenerArn: aws.String(listenerArn),
	}

	log.Debugf("Searching for elb rules for ARN '%s'", listenerArn)

	output, err := elbAPI.DescribeRules(params)
	if err != nil {
		return nil, err
	}

	rules := make([]ElbRule, len(output.Rules))
	for i, rule := range output.Rules {
		rules[i] = rule
	}

	return rules, nil
}
