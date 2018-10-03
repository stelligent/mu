package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResource(t *testing.T) {
	deploymentResource :=
		`
---
apiVersion: apps/v1beta2
kind: Deployment
metadata:
  name: hello-go
  namespace: hello-go
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hello-go
  template:
    metadata:
      labels:
        app: hello-go
    spec:
      containers:
      - name: hello-go
        image: 884669789531.dkr.ecr.us-west-2.amazonaws.com/mu-hello-go:d7f535b
        ports:
        - containerPort: 8080
`
	assert := assert.New(t)

	resourceStub, err := newResourceStub(deploymentResource)

	assert.Nil(err)
	assert.Equal("Deployment", resourceStub.Kind)
	assert.Equal("apps/v1beta2", resourceStub.APIVersion)
	assert.Equal("hello-go", resourceStub.Metadata.Name)
	assert.Equal("hello-go", resourceStub.Metadata.Namespace)

}
