package aws

import (
	"context"
	"encoding/base64"
	"os"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/ericchiang/k8s"
	logging "github.com/op/go-logging"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	yaml "gopkg.in/yaml.v2"
)

const (
	v1Prefix        = "k8s-aws-v1."
	clusterIDHeader = "x-k8s-aws-id"
)

type eksKubernetesResourceManagerProvider struct {
	eksAPI eksiface.EKSAPI
	stsAPI *sts.STS
}

type eksKubernetesResourceManager struct {
	client            *k8s.Client
	extensionsManager common.ExtensionsManager
	dryrunPath        string
}

func newEksKubernetesResourceManagerProvider(sess *session.Session) (common.KubernetesResourceManagerProvider, error) {
	log.Debug("Connecting to EKS service")
	eksAPI := eks.New(sess)
	stsAPI := sts.New(sess)

	return &eksKubernetesResourceManagerProvider{
		eksAPI: eksAPI,
		stsAPI: stsAPI,
	}, nil
}

// GetClient get a connection to eks cluster
func (eksMgrProvider *eksKubernetesResourceManagerProvider) GetResourceManager(name string) (common.KubernetesResourceManager, error) {
	eksAPI := eksMgrProvider.eksAPI
	stsAPI := eksMgrProvider.stsAPI

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

	client, err := k8s.NewClient(&k8s.Config{
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

	if err != nil {
		return nil, err
	}

	return &eksKubernetesResourceManager{
		client: client,
	}, nil
}

// UpsertResource for create/update of resources in k8s cluster
func (eksMgr *eksKubernetesResourceManager) UpsertResource(ctx context.Context,
	resource k8s.Resource,
	templateName string,
	templateData interface{}) error {

	resourceNamespace := resource.GetMetadata().GetNamespace()
	resourceName := resource.GetMetadata().GetName()

	// load existing resource for eks
	exists := true
	err := eksMgr.client.Get(ctx, resourceNamespace, resourceName, resource)
	if apiErr, ok := err.(*k8s.APIError); ok {
		if apiErr.Code == 404 {
			exists = false
		} else {
			return err
		}
	}

	// TODO: build a name for the resource
	resourceURN := "foo-bar-baz"

	// apply new values
	templateBodyReader, err := templates.NewTemplate(templateName, templateData)
	if err != nil {
		return err
	}
	templateBodyReader, err = eksMgr.extensionsManager.DecorateStackTemplate(templateName, resourceURN, templateBodyReader)
	if err != nil {
		return err
	}
	if err := yaml.NewDecoder(templateBodyReader).Decode(resource); err != nil {
		return err
	}

	// TODO: support dry-run mode
	if exists {
		log.Noticef("Updating kubernetes '%s' resource '%s' in namespace '%s' ...", reflect.TypeOf(resource), resourceName, resourceNamespace)
		if log.IsEnabledFor(logging.DEBUG) {
			yaml.NewEncoder(os.Stdout).Encode(resource)
		}
		if err := eksMgr.client.Update(ctx, resource); err != nil {
			return err
		}
	} else {
		log.Noticef("Creating kubernetes '%s' resource '%s' in namespace '%s' ...", reflect.TypeOf(resource), resourceName, resourceNamespace)
		if log.IsEnabledFor(logging.DEBUG) {
			yaml.NewEncoder(os.Stdout).Encode(resource)
		}
		if err := eksMgr.client.Create(ctx, resource); err != nil {
			return err
		}
	}

	return nil
}

// List resources in k8s cluster
func (eksMgr *eksKubernetesResourceManager) ListResources(ctx context.Context,
	namespace string,
	resourceList k8s.ResourceList) error {

	return eksMgr.client.List(ctx, namespace, resourceList)
}
