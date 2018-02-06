package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/aws/aws-sdk-go/service/s3"
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

// ShowResources shows the resources in an array
func ShowResources(output *cloudformation.DescribeStackResourcesOutput) {
	if output.StackResources != nil && len(output.StackResources) > 0 {
		stackName := output.StackResources[0].StackName
		log.Debugf("    stack %s %d resources attached", stackName, len(output.StackResources))
		for idx, resource := range output.StackResources {
			log.Debugf("   %3d: %s (%s)", idx, aws.StringValue(resource.LogicalResourceId), aws.StringValue(resource.PhysicalResourceId))
		}
	}
}

// DeleteAnyS3Buckets deletes any resources of type AWS::S3::Bucket
func DeleteAnyS3Buckets(cfnMgr *cloudformationStackManager, resources []*cloudformation.StackResource) {
	for _, resource := range resources {
		if *resource.ResourceType == "AWS::S3::Bucket" {
			fqBucketName := aws.StringValue(resource.PhysicalResourceId)
			log.Debugf("delete bucket: fullname=%s", aws.String(fqBucketName))

			// empty the bucket first
			err := EmptyS3Bucket(cfnMgr, fqBucketName)
			if err != nil {
				log.Error("couldn't delete files from bucket %s", fqBucketName)
			}
		}
	}
}

// DeleteAnyEcrRepos deletes any resources of type AWS::ECR::Repository
func DeleteAnyEcrRepos(cfnMgr *cloudformationStackManager, resources []*cloudformation.StackResource) {
	// do pre-delete API calls here (like deleting files from S3 bucket, before trying to delete bucket)
	for _, resource := range resources {
		if *resource.ResourceType == "AWS::ECR::Repository" {
			log.Debugf("ECR::Repository %V", aws.String(*resource.PhysicalResourceId))
			// TODO  - implement the following method
			err := DeleteImagesFromEcrRepo(cfnMgr.ecrAPI, *resource.PhysicalResourceId)
			if err != nil {
				log.Error("couldn't delete images from EcrRepo %s", *resource.PhysicalResourceId)
			}
		}
	}
}

// DeleteS3Bucket deletes a particular bucket
func (cfnMgr *cloudformationStackManager) DeleteS3Bucket(bucketName string) error {
	s3API := cfnMgr.s3API

	if cfnMgr.dryrunPath != "" {
		log.Infof("  DRYRUN: Skipping delete of bucket named '%s'", bucketName)
		return nil
	}
	log.Debugf("Deleting bucket named '%s'", bucketName)

	_, err := s3API.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(bucketName)})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Errorf("Delete Bucket %s %v", bucketName, aerr.Error())
		} else {
			log.Errorf("Delete Bucket %s %v", bucketName, err)
		}
		return nil
	}
	return err
}

// EmptyS3Bucket deletes the files from an S3 bucket
func EmptyS3Bucket(cfnMgr *cloudformationStackManager, bucketName string) error {
	s3API := cfnMgr.s3API
	log.Infof("s3ArtifactManager.EmptyArtifact called for bucket %s", bucketName)

	hasMoreObjects := true
	// Keep track of how many objects we delete
	totalObjects := 0

	for hasMoreObjects {
		resp, err := s3API.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucketName)})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				log.Errorf("DeleteS3BucketObjects %s %v", bucketName, aerr.Error())
			} else {
				log.Errorf("DeleteS3BucketObjects %s %v", bucketName, err)
			}
			return err
		}

		numObjs := len(resp.Contents)
		totalObjects += numObjs

		// Create Delete object with slots for the objects to delete
		var items s3.Delete
		var objs = make([]*s3.ObjectIdentifier, numObjs)

		for i, o := range resp.Contents {
			// Add objects from command line to array
			objs[i] = &s3.ObjectIdentifier{Key: aws.String(*o.Key)}
		}

		// Add list of objects to delete to Delete object
		items.SetObjects(objs)

		// Delete the items
		_, err = s3API.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &items})
		if err != nil {
			log.Errorf("Unable to delete objects from bucket %q, %v", bucketName, err)
			return err
		}

		hasMoreObjects = *resp.IsTruncated
	}

	log.Debugf("Deleted", totalObjects, "object(s) from bucket", bucketName)
	return nil
}
