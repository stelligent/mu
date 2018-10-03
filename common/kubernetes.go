package common

import (
	"context"
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
	UpsertResources(ctx context.Context, templateName string, templateData interface{}) error
}

// KubernetesResourceLister for listing kubernetes resources
type KubernetesResourceLister interface {
	ListResources(ctx context.Context, namespace string, kind string) error
}
