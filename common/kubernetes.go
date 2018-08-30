package common

import "github.com/ericchiang/k8s"

// KubernetesClientProvider for providing kubernetes client
type KubernetesClientProvider interface {
	GetClient(name string) (*k8s.Client, error)
}

// KubernetesManager for managing kubernetes
type KubernetesManager interface {
	KubernetesClientProvider
}
