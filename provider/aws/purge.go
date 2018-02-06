package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

// DeleteImagesFromEcrRepo deletes all the Docker images from a repo (so the repo itself can be deleted)
func DeleteImagesFromEcrRepo(ecrAPI ecriface.ECRAPI, repoName string) error {

	hasMoreObjects := true
	totalObjects := 0
	totalFailures := 0

	var nextToken *string
	for hasMoreObjects {
		// find all the images
		resp, err := ecrAPI.ListImages(&ecr.ListImagesInput{RepositoryName: aws.String(repoName), NextToken: nextToken})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				log.Errorf("ListImagesInput %s %v", repoName, aerr.Error())
			} else {
				log.Errorf("ListImagesInput %s %v", repoName, err)
			}
			return err
		}

		// delete them all
		result, err := ecrAPI.BatchDeleteImage(&ecr.BatchDeleteImageInput{ImageIds: resp.ImageIds, RepositoryName: &repoName})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				log.Errorf("BatchDeleteImage %v", aerr.Error())
			} else {
				log.Errorf("BatchDeleteImage %v", err)
			}
		}
		numImages := len(resp.ImageIds)
		numFailures := len(result.Failures)
		log.Debugf("%d images submitted for deletion, %d failed", numImages, numFailures)

		totalObjects += numImages
		totalFailures += numFailures
		nextToken = resp.NextToken
		hasMoreObjects = nextToken != nil
	}
	log.Debugf("total number of images found: %d, number of failed deletes %d", totalObjects, totalFailures)

	return nil
}
