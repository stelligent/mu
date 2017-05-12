package workflows

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stelligent/mu/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
)

func TestServiceLoader_FromConfig(t *testing.T) {
	assert := assert.New(t)

	ctx := new(common.Context)
	ctx.Config.Repo.Name = "myrepo"
	ctx.Config.Repo.Slug = "foo/myrepo"
	ctx.Config.Repo.Revision = "1.0.0"
	ctx.Config.Service.Name = "myservice"

	workflow := new(serviceWorkflow)
	err := workflow.serviceLoader(ctx, "2.0.0")()
	assert.Nil(err)
	assert.Equal("myservice", workflow.serviceName)
	assert.Equal("2.0.0", workflow.serviceTag)
}

func TestServiceLoader_FromRepo(t *testing.T) {
	assert := assert.New(t)

	ctx := new(common.Context)
	ctx.Config.Repo.Name = "myrepo"
	ctx.Config.Repo.Slug = "foo/myrepo"
	ctx.Config.Repo.Revision = "1.0.0"

	workflow := new(serviceWorkflow)
	err := workflow.serviceLoader(ctx, "")()
	assert.Nil(err)
	assert.Equal("myrepo", workflow.serviceName)
	assert.Equal("1.0.0", workflow.serviceTag)
}

type mockedRepositoryAuthenticator struct {
	mock.Mock
}

func (m *mockedRepositoryAuthenticator) AuthenticateRepository(repoURL string) (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestServiceRegistryAuthenticator(t *testing.T) {
	assert := assert.New(t)

	authn := new(mockedRepositoryAuthenticator)
	authn.On("AuthenticateRepository").Return(base64.StdEncoding.EncodeToString([]byte("user:pass")), nil)

	workflow := new(serviceWorkflow)
	err := workflow.serviceRegistryAuthenticator(authn)()

	assert.Nil(err)
	assert.NotNil(workflow.registryAuth)

	authJSON, err := base64.StdEncoding.DecodeString(workflow.registryAuth)
	assert.Nil(err)
	assert.Equal("{\"username\":\"user\", \"password\":\"pass\"}", string(authJSON))

	authn.AssertExpectations(t)
	authn.AssertNumberOfCalls(t, "AuthenticateRepository", 1)
}

type mockedStackManagerForService struct {
	mock.Mock
}

func (m *mockedStackManagerForService) AwaitFinalStatus(stackName string) *common.Stack {
	args := m.Called(stackName)
	stack := args.Get(0)
	if stack == nil {
		return nil
	}
	return stack.(*common.Stack)
}
func (m *mockedStackManagerForService) UpsertStack(stackName string, templateBodyReader io.Reader, stackParameters map[string]string, stackTags map[string]string) error {
	args := m.Called(stackName)
	return args.Error(0)
}
func (m *mockedStackManagerForService) DeleteStack(stackName string) error {
	args := m.Called(stackName)
	return args.Error(0)
}

func TestServiceRepoUpserter(t *testing.T) {
	assert := assert.New(t)

	svc := new(common.Service)

	workflow := new(serviceWorkflow)
	workflow.serviceName = "foo"

	stackManager := new(mockedStackManagerForUpsert)
	stackManager.On("AwaitFinalStatus", "mu-repo-foo").Return(&common.Stack{Status: cloudformation.StackStatusCreateComplete})
	stackManager.On("UpsertStack", "mu-repo-foo", mock.AnythingOfType("map[string]string")).Return(nil)

	err := workflow.serviceRepoUpserter(svc, stackManager, stackManager)()
	assert.Nil(err)

	stackManager.AssertExpectations(t)
	stackManager.AssertNumberOfCalls(t, "AwaitFinalStatus", 1)
	stackManager.AssertNumberOfCalls(t, "UpsertStack", 1)
}
