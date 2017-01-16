package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockedECS struct {
	mock.Mock
	ecsiface.ECSAPI
}

func (m *mockedECS) DescribeContainerInstances(input *ecs.DescribeContainerInstancesInput) (*ecs.DescribeContainerInstancesOutput, error) {
	args := m.Called()
	return args.Get(0).(*ecs.DescribeContainerInstancesOutput), args.Error(1)
}
func (m *mockedECS) ListContainerInstancesPages(input *ecs.ListContainerInstancesInput, cb func(*ecs.ListContainerInstancesOutput, bool) bool) error {
	args := m.Called(input, cb)
	return args.Error(0)
}

func TestEcsClusterManager_ListInstances(t *testing.T) {
	assert := assert.New(t)

	m := new(mockedECS)
	m.On("DescribeContainerInstances").Return(
		&ecs.DescribeContainerInstancesOutput{
			ContainerInstances: []*ecs.ContainerInstance{},
		}, nil)
	m.On("ListContainerInstancesPages", mock.AnythingOfType("*ecs.ListContainerInstancesInput"), mock.AnythingOfType("func(*ecs.ListContainerInstancesOutput, bool) bool")).
		Return(nil).
		Run(func(args mock.Arguments) {
			cb := args.Get(1).(func(*ecs.ListContainerInstancesOutput, bool) bool)
			cb(&ecs.ListContainerInstancesOutput{
				ContainerInstanceArns: []*string{
					aws.String("foobarbaz"),
				},
			}, true)
		})

	clusterManager := ecsClusterManager{
		ecsAPI: m,
	}

	instances, err := clusterManager.ListInstances("foo")
	assert.Nil(err)
	assert.NotNil(instances)

	m.AssertExpectations(t)
	m.AssertNumberOfCalls(t, "DescribeContainerInstances", 1)
	m.AssertNumberOfCalls(t, "ListContainerInstancesPages", 1)
}
