package aws

import (
	"crypto/md5"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stelligent/mu/common"
	"io"
	"net/http"
	"net/url"
	"os"
)

type s3ArtifactManager struct {
	s3API s3iface.S3API
	sess  *session.Session
}

func newArtifactManager(sess *session.Session) (common.ArtifactManager, error) {
	log.Debug("Connecting to S3 service")
	s3API := s3.New(sess)

	return &s3ArtifactManager{
		s3API: s3API,
		sess:  sess,
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

// GetArtifact get the artifact conditionally by etag.
func (s3Mgr *s3ArtifactManager) GetArtifact(uri string, etag string) (io.ReadCloser, string, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, "", err
	}

	if url.Scheme == "s3" {
		region, err := s3manager.GetBucketRegionWithClient(aws.BackgroundContext(), s3Mgr.s3API, url.Host)
		s3api := s3Mgr.s3API
		if aws.StringValue(s3Mgr.sess.Config.Region) != region {
			s3api = s3.New(s3Mgr.sess, aws.NewConfig().WithRegion(region))
		}
		input := &s3.GetObjectInput{
			Bucket:      aws.String(url.Host),
			Key:         aws.String(url.Path),
			IfNoneMatch: aws.String(etag),
		}
		resp, err := s3api.GetObject(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == "NotModified" {
					return resp.Body, etag, nil
				}
			}
			return nil, "", err
		}

		return resp.Body, aws.StringValue(resp.ETag), nil
	} else if url.Scheme == "https" || url.Scheme == "http" {
		req, err := http.NewRequest("GET", url.String(), nil)
		if err != nil {
			return nil, "", err
		}

		client := &http.Client{}
		req.Header.Add("If-None-Match", etag)
		resp, err := client.Do(req)
		if err != nil {
			return nil, "", err
		}

		if resp.StatusCode == 304 {
			return nil, etag, nil
		}

		return resp.Body, resp.Header.Get(http.CanonicalHeaderKey("etag")), nil
	} else if url.Scheme == "file" {
		newEtag, err := md5File(url.Path)
		if err != nil {
			return nil, "", err
		}

		if etag == "" || etag != newEtag {
			body, err := os.Open(url.Path)
			return body, newEtag, err
		}
		return nil, newEtag, nil
	}

	return nil, "", fmt.Errorf("unknown scheme on URL '%s'", url)
}

func md5File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
