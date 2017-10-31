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
	"strings"
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

	} else if url.Scheme == "s3" {
		bucket := url.Host
		object := url.Path

		// TODO: get and extract to temp dir
		log.Debugf("Loaded extension '%s' from bucket=%s object=%s", extID, bucket, object)
		//return fmt.Errorf("S3 not yet supported")
	} else if strings.HasPrefix(url.Scheme, "git+") {
		repo := strings.TrimPrefix(url.String(), "git+")

		commitish := url.Fragment
		if commitish != "" {
			repo = strings.TrimSuffix(repo, commitish)
		}

		repoURL, err := url.Parse(repo)
		if err != nil {
			return err
		}

		// TODO: git clone to temp dir
		log.Debugf("Loaded extension '%s' from repoURL=%s commitish=%s", extID, repoURL, commitish)
		//return fmt.Errorf("GIT not yet supported")

	} else {
		return fmt.Errorf("unable to handle url '%s'", url)
	}

	extensions = append(extensions, ext)
	return nil
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
