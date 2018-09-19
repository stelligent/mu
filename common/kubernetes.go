package common

import (
	"context"

	"github.com/ericchiang/k8s"
)

// KubernetesResourceManagerProvider for providing kubernetes client
type KubernetesResourceManagerProvider interface {
	GetResourceManager(name string) (KubernetesResourceManager, error)
}

// KubernetesResourceManager for managing kubernetes resources
type KubernetesResourceManager interface {
	KubernetesResourceUpserter
	KubernetesResourceLister
}

// KubernetesResourceUpserter for upserting kubernetes resources
type KubernetesResourceUpserter interface {
	UpsertResource(ctx context.Context, resource k8s.Resource, templateName string, templateData interface{}) error
}

// KubernetesResourceLister for listing kubernetes resources
type KubernetesResourceLister interface {
	ListResources(ctx context.Context, namespace string, resource k8s.ResourceList) error
}
