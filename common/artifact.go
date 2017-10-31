package common

import (
	"io"
)

// ArtifactCreator for creating artifacts
type ArtifactCreator interface {
	CreateArtifact(body io.ReadSeeker, destURI string) error
}

// ArtifactGetter for getting artifacts.  returns body, and optional error
type ArtifactGetter interface {
	GetArtifact(uri string) (io.ReadCloser, error)
}

// ArtifactManager composite of all artifact capabilities
type ArtifactManager interface {
	ArtifactCreator
	ArtifactGetter
}

