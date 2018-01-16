package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/stelligent/mu/common"
)

type snsManager struct {
	snsAPI snsiface.SNSAPI
}

func newSnsManager(sess *session.Session) (common.SubscriptionManager, error) {
	log.Debug("Connecting to SNS service")
	snsAPI := sns.New(sess)

	return &snsManager{
		snsAPI: snsAPI,
	}, nil
}

// CreateSubscription
func (snsMgr *snsManager) CreateSubscription(topic string, protocol string, endpoint string) error {
	snsAPI := snsMgr.snsAPI

	_, err := snsAPI.Subscribe(&sns.SubscribeInput{
		TopicArn: aws.String(topic),
		Endpoint: aws.String(endpoint),
		Protocol: aws.String(protocol),
	})
	return err
}

// GetSubscription
func (snsMgr *snsManager) GetSubscription(topic string, protocol string, endpoint string) (interface{}, error) {
	snsAPI := snsMgr.snsAPI

	out, err := snsAPI.ListSubscriptionsByTopic(&sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(topic),
	})

	for _, sub := range out.Subscriptions {
		if aws.StringValue(sub.Protocol) == protocol && aws.StringValue(sub.Endpoint) == endpoint {
			return sub, nil
		}
	}

	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("unable to find subscription")
}
