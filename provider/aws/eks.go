package aws

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/stelligent/mu/common"
	"github.com/stelligent/mu/templates"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
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
	client            dynamic.Interface
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

func dryRunRestConfig() dynamic.Interface {
	restConfig, _ := dynamic.NewForConfig(&rest.Config{
		Host: "",
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   []byte(""),
			Insecure: false,
		},
		BearerToken: "",
	})
	return restConfig
}

// GetResourceManager get a connection to eks cluster
func (eksMgrProvider *eksKubernetesResourceManagerProvider) GetResourceManager(name string) (common.KubernetesResourceManager, error) {
	eksAPI := eksMgrProvider.eksAPI
	stsAPI := eksMgrProvider.stsAPI

	if eksMgrProvider.dryrunPath != "" {
		return &eksKubernetesResourceManager{
			name:              name,
			client:            dryRunRestConfig(),
			dryrunPath:        eksMgrProvider.dryrunPath,
			extensionsManager: eksMgrProvider.extensionsManager,
		}, nil
	}

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

	k8sClientConfig := &rest.Config{
		Host: aws.StringValue(resp.Cluster.Endpoint),
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   certData,
			Insecure: false,
		},
		BearerToken: token,
	}

	client, err := dynamic.NewForConfig(k8sClientConfig)
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
func (eksMgr *eksKubernetesResourceManager) UpsertResources(templateName string,
	templateData interface{}) error {

	if eksMgr.dryrunPath != "" {
		return nil
	}

	// apply new values
	templateBody, err := templates.GetAsset(templateName,
		templates.DecorateTemplate(eksMgr.extensionsManager, ""),
		templates.ExecuteTemplate(templateData))

	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(templateBody))
	var b strings.Builder
	for scanner.Scan() {
		if scanner.Text() == "---" {
			// flush current resource
			if b.Len() > 0 {
				err = eksMgr.upsertResource(b.String())
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
		err = eksMgr.upsertResource(b.String())
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

func (eksMgr *eksKubernetesResourceManager) upsertResource(resourceBody string) error {
	stub, err := newResourceStub(resourceBody)
	if err != nil {
		return err
	}

	resourceKind := stub.Kind
	resourceNamespace := stub.Metadata.Namespace
	resourceName := stub.Metadata.Name

	resourceClient, err := eksMgr.getResourceInterface(stub.APIVersion, stub.Kind)
	if err != nil {
		return nil
	}

	// load existing resource for eks
	exists := true
	options := metav1.GetOptions{}
	_, err = resourceClient.Namespace(resourceNamespace).Get(resourceName, options)

	if err != nil {
		if errors.IsNotFound(err) {
			exists = false
		} else {
			return err
		}
	}

	resourceURN := fmt.Sprintf("%s-%s-%s", eksMgr.name, resourceKind, resourceName)
	resourceBody, err = templates.DecorateTemplate(eksMgr.extensionsManager, resourceURN)("", resourceBody)
	if err != nil {
		return err
	}

	resource := &unstructured.Unstructured{Object: make(map[string]interface{})}
	yaml.Unmarshal([]byte(resourceBody), resource.Object)
	resource.Object = common.ConvertMapI2MapS(resource.Object).(map[string]interface{})

	if eksMgr.dryrunPath != "" {
		err := writeResource(eksMgr.dryrunPath, resourceURN, resourceBody)
		if err != nil {
			return err
		}
		var action string
		if exists {
			action = "update"
		} else {
			action = "create"
		}
		log.Infof("  DRYRUN: Skipping %s of resource named '%s'.  File written to '%s'", action, resourceURN, eksMgr.dryrunPath)
		return nil
	}

	if exists {
		log.Infof("  Patching namespace:%s type:%s name:%s", resourceNamespace, resourceKind, resourceName)
		s, err := json.Marshal(resource.UnstructuredContent())
		if err != nil {
			return err
		}
		if _, err := resourceClient.Namespace(resourceNamespace).Patch(resourceName, types.StrategicMergePatchType, s); err != nil {
			return err
		}
	} else {
		log.Infof("  Creating namespace:%s type:%s name:%s", resourceNamespace, resourceKind, resourceName)
		if _, err := resourceClient.Namespace(resourceNamespace).Create(resource); err != nil {
			return err
		}
	}

	return nil
}

// DeleteResource for deletion of resources in k8s cluster
func (eksMgr *eksKubernetesResourceManager) DeleteResource(apiVersion string, kind string, namespace string, name string) error {
	resourceClient, err := eksMgr.getResourceInterface(apiVersion, kind)
	if err != nil {
		return nil
	}
	return resourceClient.Namespace(namespace).Delete(name, nil)
}

// List resources in k8s cluster
func (eksMgr *eksKubernetesResourceManager) ListResources(apiVersion string, kind string, namespace string) (*unstructured.UnstructuredList, error) {
	resourceClient, err := eksMgr.getResourceInterface(apiVersion, kind)
	if err != nil {
		return nil, nil
	}
	return resourceClient.Namespace(namespace).List(metav1.ListOptions{})
}

func (eksMgr *eksKubernetesResourceManager) getResourceInterface(apiVersion string, kind string) (dynamic.NamespaceableResourceInterface, error) {
	groupVersion, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}
	groupVersionKind := groupVersion.WithKind(kind)
	resourceGroupVersion, _ := meta.UnsafeGuessKindToResource(groupVersionKind)
	return eksMgr.client.Resource(resourceGroupVersion), nil
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
