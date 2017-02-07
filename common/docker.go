package common

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"io"
	"os"
	"path/filepath"
)

// DockerImageBuilder for creating docker images
type DockerImageBuilder interface {
	ImageBuild(contextDir string, relDockerfile string, tags []string, dockerOut io.Writer) error
}

// DockerImagePusher for pushing docker images
type DockerImagePusher interface {
	ImagePush(image string, registryAuth string, dockerOut io.Writer) error
}

// DockerManager composite of all cluster capabilities
type DockerManager interface {
	DockerImageBuilder
	DockerImagePusher
}

type clientDockerManager struct {
	dockerClient *client.Client
}

func newClientDockerManager() (DockerManager, error) {
	log.Debug("Connecting to Docker daemon")
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return &clientDockerManager{
		dockerClient: cli,
	}, nil
}

func (d *clientDockerManager) ImageBuild(contextDir string, relDockerfile string, tags []string, dockerOut io.Writer) error {
	options := types.ImageBuildOptions{
		Tags: tags,
	}

	buildContext, err := createBuildContext(contextDir, relDockerfile)
	if err != nil {
		return err
	}

	defer buildContext.Close()

	log.Debugf("Creating image from context dir '%s' with tag '%s'", contextDir, tags)
	resp, err := d.dockerClient.ImageBuild(context.Background(), buildContext, options)
	if err != nil {
		return err
	}

	return handleDockerResponse(resp.Body, dockerOut)
}

func createBuildContext(contextDir string, relDockerfile string) (io.ReadCloser, error) {
	log.Debugf("Creating archive for build context dir '%s' with relative dockerfile '%s'", contextDir, relDockerfile)

	// And canonicalize dockerfile name to a platform-independent one
	relDockerfile, err := archive.CanonicalTarNameForPath(relDockerfile)
	if err != nil {
		return nil, fmt.Errorf("cannot canonicalize dockerfile path %s: %v", relDockerfile, err)
	}

	f, err := os.Open(filepath.Join(contextDir, ".dockerignore"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	defer f.Close()

	var excludes []string
	if err == nil {
		excludes, err = dockerignore.ReadAll(f)
		if err != nil {
			return nil, err
		}
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed. The daemon will remove them for us, if needed, after it
	// parses the Dockerfile. Ignore errors here, as they will have been
	// caught by validateContextDirectory above.
	var includes = []string{"."}
	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches(relDockerfile, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", relDockerfile)
	}

	compression := archive.Uncompressed
	buildCtx, err := archive.TarWithOptions(contextDir, &archive.TarOptions{
		Compression:     compression,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	})
	if err != nil {
		return nil, err
	}

	return buildCtx, nil
}

func (d *clientDockerManager) ImagePush(image string, registryAuth string, dockerOut io.Writer) error {

	log.Debugf("Pushing image '%s' auth '%s'", image, registryAuth)

	pushOptions := types.ImagePushOptions{
		RegistryAuth: registryAuth,
	}

	resp, err := d.dockerClient.ImagePush(context.Background(), image, pushOptions)
	if err != nil {
		return err
	}

	return handleDockerResponse(resp, dockerOut)
}

type dockerMessage struct {
	ID          string `json:"id"`
	Stream      string `json:"stream"`
	Error       string `json:"error"`
	ErrorDetail struct {
		Message string
	}
	Status   string `json:"status"`
	Progress string `json:"progress"`
}

func handleDockerResponse(resp io.ReadCloser, dockerOut io.Writer) error {
	defer resp.Close()

	scanner := bufio.NewScanner(resp)
	msg := dockerMessage{}
	for scanner.Scan() {
		line := scanner.Bytes()
		msg.ID = ""
		msg.Stream = ""
		msg.Error = ""
		msg.ErrorDetail.Message = ""
		msg.Status = ""
		msg.Progress = ""
		if err := json.Unmarshal(line, &msg); err == nil {
			if msg.Error != "" {
				return fmt.Errorf("%s", msg.Error)
			} else if dockerOut != nil {
				if msg.Status != "" {
					if msg.Progress != "" {
						dockerOut.Write([]byte(fmt.Sprintf("  %s :: %s :: %s\n", msg.Status, msg.ID, msg.Progress)))
					} else {
						dockerOut.Write([]byte(fmt.Sprintf("  %s :: %s\n", msg.Status, msg.ID)))
					}
				} else if msg.Stream != "" {
					dockerOut.Write([]byte(fmt.Sprintf("  %s", msg.Stream)))
				} else {
					log.Debugf("Unable to handle line: %s", string(line))
				}
			}
		} else {
			log.Debugf("Unable to unmarshal line [%s] ==> %v", string(line), err)
		}
	}

	return nil
}
