package common

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// KubernetesResourceManagerProvider for providing kubernetes client
type KubernetesResourceManagerProvider interface {
	GetResourceManager(name string) (KubernetesResourceManager, error)
}

// KubernetesResourceManager for managing kubernetes resources
type KubernetesResourceManager interface {
	KubernetesResourceUpserter
	KubernetesResourceLister
	KubernetesResourceDeleter
}

// KubernetesResourceUpserter for upserting kubernetes resources
type KubernetesResourceUpserter interface {
	UpsertResources(templateName string, templateData interface{}) error
}

// KubernetesResourceLister for listing kubernetes resources
type KubernetesResourceLister interface {
	ListResources(apiVersion string, kind string, namespace string) (*unstructured.UnstructuredList, error)
}

// KubernetesResourceDeleter for deleting kubernetes resources
type KubernetesResourceDeleter interface {
	DeleteResource(apiVersion string, kind string, namespace string, name string) error
}
