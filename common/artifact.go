package common

import (
	"io"
)

// ArtifactCreator for creating artifacts
type ArtifactCreator interface {
	CreateArtifact(body io.ReadSeeker, destURI string, kmsKey string) error
}

// BucketEmptier for emptying buckets
type BucketEmptier interface {
	EmptyBucket(bucketName string) error
}

// ArtifactGetter for getting artifacts.  conditional get (based on etag).  returns body, etag and optional error
type ArtifactGetter interface {
	GetArtifact(uri string, etag string) (io.ReadCloser, string, error)
}

// ArtifactManager composite of all artifact capabilities
type ArtifactManager interface {
	ArtifactCreator
	ArtifactGetter
	BucketEmptier
}
