package common

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/mholt/archiver"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type extension struct {
	id   string
	path string
	etag string
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

	var ext = &extension{
		extID,
		extensionDirectory,
		"",
	}

	// check for existing etag
	etagBytes, err := ioutil.ReadFile(filepath.Join(extensionDirectory, ".etag"))
	if err == nil {
		ext.etag = string(etagBytes)
	}

	if fi, err := os.Stat(url.Path); url.Scheme == "file" && err == nil && fi.IsDir() {
		ext.path = url.Path
		log.Debugf("Loaded extension from '%s'", url.Path)
	} else {
		body, etag, err := ctx.ArtifactManager.GetArtifact(url.String(), ext.etag)
		if err != nil {
			return err
		}

		if body != nil {
			defer body.Close()

			// empty dir
			os.RemoveAll(extensionDirectory)
			os.MkdirAll(extensionDirectory, 0700)

			// write out archive to dir
			err = extractArchive(ext.path, body)
			if err != nil {
				return err
			}

			// write new etag
			err = ioutil.WriteFile(filepath.Join(extensionDirectory, ".etag"), []byte(etag), 0644)
			if err != nil {
				return err
			}
			log.Debugf("Loaded extension from '%s' [id=%s]", url, extID)
		} else {
			log.Debugf("Loaded extension from cache [id=%s]", extID)
		}

	}

	extManifest := make(map[interface{}]interface{})
	extManifestFile, err := ioutil.ReadFile(filepath.Join(ext.path, "mu-extension.yml"))
	if err == nil {
		err = yaml.Unmarshal(extManifestFile, extManifest)
		if err != nil {
			log.Debugf("error unmarshalling mu-extension.yml: %s", err)
		}
	} else {
		log.Debugf("error reading mu-extension.yml: %s", err)
	}

	if name, ok := extManifest["name"]; ok {
		if version, ok := extManifest["version"]; ok {
			log.Infof("Loaded extension %s (version=%v)", name, version)
		} else {
			log.Infof("Loaded extension %s", name)
		}
	} else {
		log.Infof("Loaded extension %s", url)
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
