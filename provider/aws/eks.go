package aws

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/ericchiang/k8s"
	"github.com/golang/protobuf/proto"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	yaml "gopkg.in/yaml.v2"
)

const (
	v1Prefix        = "k8s-aws-v1."
	clusterIDHeader = "x-k8s-aws-id"
)

type eksKubernetesResourceManagerProvider struct {
	eksAPI            eksiface.EKSAPI
	stsAPI            *sts.STS
	extensionsManager common.ExtensionsManager
	dryrunPath        string
}

type eksKubernetesResourceManager struct {
	name              string
	client            *k8s.Client
	extensionsManager common.ExtensionsManager
	dryrunPath        string
}

func newEksKubernetesResourceManagerProvider(sess *session.Session, extensionsManager common.ExtensionsManager, dryrunPath string) (common.KubernetesResourceManagerProvider, error) {
	if dryrunPath != "" {
		log.Debugf("Running in DRYRUN mode with path '%s'", dryrunPath)
	}
	log.Debug("Connecting to EKS service")
	eksAPI := eks.New(sess)
	stsAPI := sts.New(sess)

	return &eksKubernetesResourceManagerProvider{
		eksAPI:            eksAPI,
		stsAPI:            stsAPI,
		dryrunPath:        dryrunPath,
		extensionsManager: extensionsManager,
	}, nil
}

// GetResourceManager get a connection to eks cluster
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
		name:              name,
		client:            client,
		dryrunPath:        eksMgrProvider.dryrunPath,
		extensionsManager: eksMgrProvider.extensionsManager,
	}, nil
}

// UpsertResources for create/update of resources in k8s cluster
func (eksMgr *eksKubernetesResourceManager) UpsertResources(ctx context.Context,
	templateName string,
	templateData interface{}) error {

	// apply new values
	templateBody, err := templates.GetAsset(templateName,
		templates.ExecuteTemplate(templateData),
		templates.DecorateTemplate(eksMgr.extensionsManager, ""))

	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(templateBody))
	var b strings.Builder
	for scanner.Scan() {
		if scanner.Text() == "---" {
			// flush current resource
			if b.Len() > 0 {
				err = eksMgr.upsertResource(ctx, b.String())
				if err != nil {
					return err
				}
				b.Reset()
			}
		} else {
			// append to current resource
			_, err = fmt.Fprintln(&b, scanner.Text())
			if err != nil {
				return err
			}
		}
	}

	if b.Len() > 0 {
		err = eksMgr.upsertResource(ctx, b.String())
		if err != nil {
			return err
		}
	}
	return nil
}

func newResourceStub(resourceBody string) (*resourceStub, error) {
	stub := &resourceStub{}
	err := yaml.Unmarshal([]byte(resourceBody), stub)
	if err != nil {
		return nil, err
	}
	return stub, nil
}

type resourceStub struct {
	APIVersion string `yaml:"apiVersion,omitempty"`
	Kind       string
	Metadata   struct {
		Name      string
		Namespace string
	}
}

func (stub *resourceStub) newResource() (k8s.Resource, error) {
	versionParts := strings.Split(stub.APIVersion, "/")
	var kind string
	if len(versionParts) == 1 {
		kind = fmt.Sprintf("k8s.io.api.core.%s.%s", stub.APIVersion, stub.Kind)
	} else {
		kind = fmt.Sprintf("k8s.io.api.%s.%s", strings.Join(versionParts, "."), stub.Kind)
	}
	resourceType := proto.MessageType(kind)
	if resourceType == nil {
		return nil, fmt.Errorf("Unable to determine type for %v", kind)
	}
	resourceType = resourceType.Elem()
	log.Debugf("kind: %v -> %v", kind, resourceType)
	r := reflect.New(resourceType)
	return r.Interface().(k8s.Resource), nil
}

func (eksMgr *eksKubernetesResourceManager) upsertResource(ctx context.Context, resourceBody string) error {
	stub, err := newResourceStub(resourceBody)
	if err != nil {
		return err
	}

	resource, err := stub.newResource()
	if err != nil {
		return err
	}

	resourceType := stub.Kind
	resourceNamespace := stub.Metadata.Namespace
	resourceName := stub.Metadata.Name

	// load existing resource for eks
	exists := true
	err = eksMgr.client.Get(ctx, resourceNamespace, resourceName, resource)
	if apiErr, ok := err.(*k8s.APIError); ok {
		if apiErr.Code == 404 {
			exists = false
		} else {
			return err
		}
	}

	resourceURN := fmt.Sprintf("%s-%s-%s", eksMgr.name, resourceType, resourceName)
	resourceBody, err = templates.DecorateTemplate(eksMgr.extensionsManager, resourceURN)("", resourceBody)
	if err != nil {
		return err
	}

	if err := yaml.NewDecoder(strings.NewReader(resourceBody)).Decode(resource); err != nil {
		return err
	}

	if eksMgr.dryrunPath != "" {
		err := writeResource(eksMgr.dryrunPath, resourceURN, resourceBody)
		if err != nil {
			return err
		}
		log.Infof("  DRYRUN: Skipping create of resource named '%s'.  File written to '%s'", resourceURN, eksMgr.dryrunPath)
		return nil
	}

	if exists {
		log.Noticef("Updating kubernetes '%s' resource '%s' in namespace '%s' ...", resourceType, resourceName, resourceNamespace)
		if err := eksMgr.client.Update(ctx, resource); err != nil {
			return err
		}
	} else {
		log.Noticef("Creating kubernetes '%s' resource '%s' in namespace '%s' ...", resourceType, resourceName, resourceNamespace)
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

func writeResource(directory string, resourceName string, resourceBody string) error {
	os.MkdirAll(directory, 0700)
	resourceFile := fmt.Sprintf("%s/resource-%s.yml", directory, resourceName)
	fileWriter, err := os.Create(resourceFile)
	if err != nil {
		return err
	}
	defer fileWriter.Close()
	l, err := fileWriter.WriteString(resourceBody)
	log.Debugf("Wrote %d bytes to %s", l, resourceFile)
	return err
}
