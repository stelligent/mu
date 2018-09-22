package workflows

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/stelligent/mu/common"
)

// NewServicePusher create a new workflow for pushing a service to a repo
func NewServicePusher(ctx *common.Context, tag string, provider string, kmsKey string, dockerWriter io.Writer) Executor {

	workflow := new(serviceWorkflow)

	return newPipelineExecutor(
		workflow.serviceLoader(ctx, tag, provider),
		newConditionalExecutor(workflow.isEcrProvider(),
			newPipelineExecutor(
				workflow.serviceRepoUpserter(ctx.Config.Namespace, &ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceImageBuilder(ctx.DockerManager, &ctx.Config, dockerWriter),
				workflow.serviceRegistryAuthenticator(ctx.ClusterManager),
				workflow.serviceImagePusher(ctx.DockerManager, dockerWriter),
			),
			newPipelineExecutor(
				workflow.serviceBucketUpserter(ctx.Config.Namespace, &ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceArchiveUploader(ctx.Config.Basedir, ctx.ArtifactManager, kmsKey),
			)))

}

func (workflow *serviceWorkflow) serviceImageBuilder(imageBuilder common.DockerImageBuilder, config *common.Config, dockerWriter io.Writer) Executor {
	return func() error {
		log.Noticef("Building service:'%s' as image:%s'", workflow.serviceName, workflow.serviceImage)
		return imageBuilder.ImageBuild(config.Basedir, workflow.serviceName, config.Service.Dockerfile, []string{workflow.serviceImage}, dockerWriter)
	}
}

func (workflow *serviceWorkflow) serviceImagePusher(imagePusher common.DockerImagePusher, dockerWriter io.Writer) Executor {
	return func() error {
		log.Noticef("Pushing service '%s' to '%s'", workflow.serviceName, workflow.serviceImage)
		return imagePusher.ImagePush(workflow.serviceImage, workflow.registryAuth, dockerWriter)
	}
}

func (workflow *serviceWorkflow) serviceArchiveUploader(basedir string, artifactCreator common.ArtifactCreator, kmsKey string) Executor {
	return func() error {
		destURL := fmt.Sprintf("s3://%s/%s", workflow.appRevisionBucket, workflow.appRevisionKey)
		log.Noticef("Pushing archive '%s' to '%s'", basedir, destURL)

		zipfile, err := zipDir(fmt.Sprintf("%s/", basedir))
		if err != nil {
			return err
		}
		defer os.Remove(zipfile.Name()) // clean up

		err = artifactCreator.CreateArtifact(zipfile, destURL, kmsKey)
		if err != nil {
			return err
		}

		return nil
	}
}

func zipDir(basedir string) (*os.File, error) {
	zipfile, err := ioutil.TempFile("", "artifact")
	if err != nil {
		return nil, err
	}

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	log.Debugf("Creating zipfile '%s' from basedir '%s'", zipfile.Name(), basedir)

	filepath.Walk(basedir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, basedir)

		if header.Name == "" {
			return nil
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			log.Debugf(" ..Adding file '%s'", header.Name)
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return zipfile, err
}
