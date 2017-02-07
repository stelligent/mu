package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
)

// ElbRuleLister for getting cluster instances
type ElbRuleLister interface {
	ListRules(listenerArn string) ([]*elbv2.Rule, error)
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
func (elbMgr *elbv2Manager) ListRules(listenerArn string) ([]*elbv2.Rule, error) {
	elbAPI := elbMgr.elbAPI

	params := &elbv2.DescribeRulesInput{
		ListenerArn: aws.String(listenerArn),
	}

	log.Debugf("Searching for elb rules for ARN '%s'", listenerArn)

	output, err := elbAPI.DescribeRules(params)
	if err != nil {
		return nil, err
	}

	return output.Rules, nil
}
