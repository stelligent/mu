package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/ericchiang/k8s"
	"github.com/stelligent/mu/common"
)

type eksKubernetesManager struct {
	eksAPI eksiface.EKSAPI
}

func newEksKubernetesManager(sess *session.Session) (common.KubernetesManager, error) {
	log.Debug("Connecting to EKS service")
	eksAPI := eks.New(sess)

	return &eksKubernetesManager{
		eksAPI: eksAPI,
	}, nil
}

// GetClient get a connection to eks cluster
func (eksMgr *eksKubernetesManager) GetClient(name string) (*k8s.Client, error) {
	eksAPI := eksMgr.eksAPI

	resp, err := eksAPI.DescribeCluster(&eks.DescribeClusterInput{
		Name: aws.String(name),
	})

	if err != nil {
		return nil, err
	}

	k8s.NewClient(&k8s.Config{
		Clusters: []k8s.NamedCluster{
			{
				Name: name,
				Cluster: k8s.Cluster{
					InsecureSkipTLSVerify: false,
					CertificateAuthority:  aws.StringValue(resp.Cluster.CertificateAuthority.Data),
					Server:                aws.StringValue(resp.Cluster.Endpoint),
				},
			},
		},
		Contexts: []k8s.NamedContext{
			k8s.NamedContext{
				Name: name,
				Context: k8s.Context{
					Cluster:  name,
					AuthInfo: "foo",
				},
			},
		},
		CurrentContext: name,
	})

	return nil, nil
}
