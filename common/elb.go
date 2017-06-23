package common

import (
	"github.com/aws/aws-sdk-go/service/elbv2"
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
