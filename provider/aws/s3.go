package aws

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stelligent/mu/common"
)

type s3ArtifactManager struct {
	s3API      s3iface.S3API
	sess       *session.Session
	dryrunPath string
}

func newArtifactManager(sess *session.Session, dryrunPath string) (common.ArtifactManager, error) {
	log.Debug("Connecting to S3 service")
	s3API := s3.New(sess)

	return &s3ArtifactManager{
		s3API:      s3API,
		sess:       sess,
		dryrunPath: dryrunPath,
	}, nil
}

// CreateArtifact get the instances for a specific cluster
func (s3Mgr *s3ArtifactManager) CreateArtifact(body io.ReadSeeker, destURL string, kmsKey string) error {
	s3API := s3Mgr.s3API

	s3URL, err := url.Parse(destURL)
	if err != nil {
		return err
	}
	if s3URL.Scheme != "s3" {
		return fmt.Errorf("destURL must have scheme of 's3', received '%s'", s3URL.Scheme)
	}

	// start from the begining
	body.Seek(0, 0)

	if s3Mgr.dryrunPath != "" {
		err := os.MkdirAll(s3Mgr.dryrunPath, 0700)
		if err != nil {
			return err
		}
		artifactFile := fmt.Sprintf("%s/%s-%s", s3Mgr.dryrunPath, s3URL.Host, path.Base(s3URL.Path))
		f, err := os.OpenFile(artifactFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, body)
		if err != nil {
			return err
		}

		log.Infof("  DRYRUN: Skipping PUT of artifact '%s'. Written to '%s'", destURL, artifactFile)
		return nil
	}

	params := &s3.PutObjectInput{
		Bucket: aws.String(s3URL.Host),
		Key:    aws.String(s3URL.Path),
		Body:   body,
	}

	if kmsKey != "" {
		params.SSEKMSKeyId = aws.String(kmsKey)
		params.ServerSideEncryption = aws.String("aws:kms")
	}

	log.Debugf("Creating artifact at '%s'", destURL)

	_, err = s3API.PutObject(params)
	if err != nil {
		return err
	}

	return nil
}

func (s3Mgr *s3ArtifactManager) getArtifactS3(url *url.URL, etag string) (io.ReadCloser, string, error) {
	region, err := s3manager.GetBucketRegionWithClient(aws.BackgroundContext(), s3Mgr.s3API, url.Host)
	if err != nil {
		return nil, "", err
	}
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
}

func (s3Mgr *s3ArtifactManager) getArtifactHTTP(url *url.URL, etag string) (io.ReadCloser, string, error) {
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
}

func (s3Mgr *s3ArtifactManager) getArtifactFile(url *url.URL, etag string) (io.ReadCloser, string, error) {
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

// GetArtifact get the artifact conditionally by etag.
func (s3Mgr *s3ArtifactManager) GetArtifact(uri string, etag string) (body io.ReadCloser, etagRet string, err error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, "", err
	}

	switch url.Scheme {
	case "s3":
		body, etagRet, err = s3Mgr.getArtifactS3(url, etag)
	case "http":
		fallthrough
	case "https":
		body, etagRet, err = s3Mgr.getArtifactHTTP(url, etag)
	case "file":
		body, etagRet, err = s3Mgr.getArtifactFile(url, etag)
	default:
		body = nil
		etagRet = ""
		err = fmt.Errorf("unknown scheme on URL '%s'", url)
	}
	return
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

// EmptyBucket to empty bucket
func (s3Mgr *s3ArtifactManager) EmptyBucket(bucketName string) error {
	if s3Mgr.dryrunPath != "" {
		log.Infof("  DRYRUN: Skipping emptying of bucket '%s'", bucketName)
		return nil
	}
	log.Infof("  Emptying bucket '%s'", bucketName)
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}
	for {
		//Requesting for batch of objects from s3 bucket
		objects, err := s3Mgr.s3API.ListObjects(params)
		if err != nil {
			return err
		}
		//Checks if the bucket is already empty
		if len((*objects).Contents) == 0 {
			log.Debug("Bucket is already empty")
			return nil
		}
		log.Debug("First object in batch | ", *(objects.Contents[0].Key))

		//creating an array of pointers of ObjectIdentifier
		objectsToDelete := make([]*s3.ObjectIdentifier, 0, 1000)
		for _, object := range (*objects).Contents {
			obj := s3.ObjectIdentifier{
				Key: object.Key,
			}
			objectsToDelete = append(objectsToDelete, &obj)
		}
		//Creating JSON payload for bulk delete
		deleteArray := s3.Delete{Objects: objectsToDelete}
		deleteParams := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &deleteArray,
		}
		//Running the Bulk delete job (limit 1000)
		_, err = s3Mgr.s3API.DeleteObjects(deleteParams)
		if err != nil {
			return err
		}
		if *(*objects).IsTruncated { //if there are more objects in the bucket, IsTruncated = true
			params.Marker = (*deleteParams).Delete.Objects[len((*deleteParams).Delete.Objects)-1].Key
			log.Debug("Requesting next batch | ", *(params.Marker))
		} else { //if all objects in the bucket have been cleaned up.
			break
		}
	}
	return nil
}
