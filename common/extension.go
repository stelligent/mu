package common

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"io"
	"bufio"
	"net/http"
	"github.com/mholt/archiver"
)

type extension struct {
	id   string
	path string
}

var extensions = make([]*extension, 0)

func urlToID(url *url.URL) string {
	h := sha1.New()
	h.Write([]byte(url.String()))
	return hex.EncodeToString(h.Sum(nil))
}

func loadExtension(ctx *Context, url *url.URL) error {
	if !url.IsAbs() {
		// Assume relative path
		return fmt.Errorf("unable to handle relative path '%s'", url)
	}

	extID := urlToID(url)

	for _, existingExt := range extensions {
		if existingExt.id == extID {
			log.Warningf("Extension '%s' already loaded...skipping.", url.String())
			return nil
		}
	}

	userdir, err := homedir.Dir()
	if err != nil {
		return err
	}

	extensionDirectory := filepath.Join(userdir, ".mu", "extensions", extID)
	os.MkdirAll(extensionDirectory, 0700)
	var ext = &extension{
		extID,
		extensionDirectory,
	}

	if url.Scheme == "file" {
		ext.path = url.Path

		if _, err := os.Stat(ext.path); err != nil {
			return err
		}

		log.Debugf("Loaded extension '%s' from path=%s", extID, ext.path)

	} else {
		body, err := ctx.ArtifactManager.GetArtifact(url.String())
		if err != nil {
			return err
		}

		defer body.Close()
		err = extractArchive(ext.path, body)
		if err != nil {
			return err
		}

		log.Debugf("Loaded extension '%s' from url=%s", extID, url)
	}

	extensions = append(extensions, ext)
	return nil
}

func extractArchive(destPath string, archive io.ReadCloser) error {
	reader := bufio.NewReader(archive)
	headBytes, err := reader.Peek(512)
	if err != nil {
		return err
	}
	contentType := http.DetectContentType(headBytes)
	log.Debugf("Extracting type '%s'", contentType)

	switch contentType {
	case "application/x-gzip":
		return archiver.TarGz.Read(reader, destPath)
	case "application/zip":
		return archiver.Zip.Read(reader, destPath)
	default:
		return fmt.Errorf("unable to handle archive of content-type '%s'", contentType)
	}

}

// GetCfnUpdatesFromExtensions finds all CFN updates across all extensions
func GetCfnUpdatesFromExtensions(assetName string) []interface{} {
	cfnUpdates := make([]interface{}, 0)
	for _, ext := range extensions {
		assetPath := filepath.Join(ext.path, assetName)
		yamlFile, err := ioutil.ReadFile(assetPath)
		if err != nil {
			log.Debugf("Unable to find asset '%s' in extension '%s': %s", assetName, ext.id, err)
		} else {
			cfnUpdate := make(map[interface{}]interface{})
			err = yaml.Unmarshal(yamlFile, cfnUpdate)
			if err != nil {
				log.Warningf("Unable to parse asset '%s' in extension '%s': %s", assetName, ext.id, err)
			} else {
				cfnUpdates = append(cfnUpdates, cfnUpdate)
			}
		}
	}
	return cfnUpdates
}
