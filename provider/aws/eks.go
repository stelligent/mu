package aws

import (
	"encoding/base64"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/ericchiang/k8s"
	"github.com/stelligent/mu/common"
)

const (
	v1Prefix         = "k8s-aws-v1."
	maxTokenLenBytes = 1024 * 4
	clusterIDHeader  = "x-k8s-aws-id"
)

type eksKubernetesManager struct {
	eksAPI eksiface.EKSAPI
	stsAPI *sts.STS
}

func newEksKubernetesManager(sess *session.Session) (common.KubernetesManager, error) {
	log.Debug("Connecting to EKS service")
	eksAPI := eks.New(sess)
	creds := stscreds.NewCredentials(sess, "arn:aws:iam::884669789531:role/eks-cloudformation-common-us-west-2")
	stsAPI := sts.New(sess, &aws.Config{Credentials: creds})

	return &eksKubernetesManager{
		eksAPI: eksAPI,
		stsAPI: stsAPI,
	}, nil
}

// GetClient get a connection to eks cluster
func (eksMgr *eksKubernetesManager) GetClient(name string) (*k8s.Client, error) {
	eksAPI := eksMgr.eksAPI
	stsAPI := eksMgr.stsAPI

	resp, err := eksAPI.DescribeCluster(&eks.DescribeClusterInput{
		Name: aws.String(name),
	})

	if err != nil {
		return nil, err
	}

	certData, err := base64.StdEncoding.DecodeString(aws.StringValue(resp.Cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}

	request, _ := stsAPI.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	request.HTTPRequest.Header.Add(clusterIDHeader, name)

	// sign the request
	presignedURLString, err := request.Presign(60 * time.Second)
	if err != nil {
		return nil, err
	}
	token := v1Prefix + base64.RawURLEncoding.EncodeToString([]byte(presignedURLString))

	return k8s.NewClient(&k8s.Config{
		Clusters: []k8s.NamedCluster{
			{
				Name: name,
				Cluster: k8s.Cluster{
					InsecureSkipTLSVerify:    false,
					CertificateAuthorityData: certData,
					Server: aws.StringValue(resp.Cluster.Endpoint),
				},
			},
		},
		Contexts: []k8s.NamedContext{
			k8s.NamedContext{
				Name: name,
				Context: k8s.Context{
					Cluster:  name,
					AuthInfo: "mu",
				},
			},
		},
		AuthInfos: []k8s.NamedAuthInfo{
			k8s.NamedAuthInfo{
				Name: "mu",
				AuthInfo: k8s.AuthInfo{
					Token: token,
				},
			},
		},
		CurrentContext: name,
	})
}
