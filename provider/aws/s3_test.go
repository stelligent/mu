package aws

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedS3 struct {
	mock.Mock
	s3iface.S3API
}

func (m *mockedS3) PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	args := m.Called()
	return nil, args.Error(1)
}

func TestS3ArtifactManager_CreateArtifact(t *testing.T) {
	assertion := assert.New(t)
	s3Mock := new(mockedS3)

	s3Mock.On("PutObject").Return(&s3.PutObjectOutput{}, nil)

	artifactManager := s3ArtifactManager{
		s3API: s3Mock,
	}

	err := artifactManager.CreateArtifact(strings.NewReader("foo"), "s3://bucket/key", "key")
	assertion.Nil(err)

	s3Mock.AssertExpectations(t)
	s3Mock.AssertNumberOfCalls(t, "PutObject", 1)
}
