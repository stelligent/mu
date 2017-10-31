package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stelligent/mu/common"
	"io"
	"net/url"
	"net/http"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3ArtifactManager struct {
	s3API s3iface.S3API
	sess *session.Session
}

func newArtifactManager(sess *session.Session) (common.ArtifactManager, error) {
	log.Debug("Connecting to S3 service")
	s3API := s3.New(sess)

	return &s3ArtifactManager{
		s3API: s3API,
		sess: sess,
	}, nil
}

// CreateArtifact get the instances for a specific cluster
func (s3Mgr *s3ArtifactManager) CreateArtifact(body io.ReadSeeker, destURL string) error {
	s3API := s3Mgr.s3API

	s3URL, err := url.Parse(destURL)
	if err != nil {
		return err
	}
	if s3URL.Scheme != "s3" {
		return fmt.Errorf("destURL must have scheme of 's3', recieved '%s'", s3URL.Scheme)
	}

	params := &s3.PutObjectInput{
		Bucket: aws.String(s3URL.Host),
		Key:    aws.String(s3URL.Path),
		Body:   body,
	}

	log.Debugf("Creating artifact at '%s'", destURL)

	_, err = s3API.PutObject(params)
	if err != nil {
		return err
	}

	return nil
}

// GetArtifact get the instances for a specific cluster
func (s3Mgr *s3ArtifactManager) GetArtifact(uri string) (io.ReadCloser, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if url.Scheme == "s3" {
		region, err := s3manager.GetBucketRegionWithClient(aws.BackgroundContext(), s3Mgr.s3API, url.Host)
		s3api := s3Mgr.s3API
		if aws.StringValue(s3Mgr.sess.Config.Region) != region {
			s3api = s3.New(s3Mgr.sess, aws.NewConfig().WithRegion(region))
		}
		input := &s3.GetObjectInput{
			Bucket: aws.String(url.Host),
			Key: aws.String(url.Path),
		}
		resp, err := s3api.GetObject(input)
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	} else if url.Scheme == "https" || url.Scheme == "http" {
		resp, err := http.Get(url.String())
		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	}

	return nil, fmt.Errorf("unknown scheme on URL '%s'", url)
}




