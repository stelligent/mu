package common

import "io"

// ArtifactCreator for getting cluster instances
type ArtifactCreator interface {
	CreateArtifact(body io.ReadSeeker, destURI string) error
}

// ArtifactManager composite of all artifact capabilities
type ArtifactManager interface {
	ArtifactCreator
}
