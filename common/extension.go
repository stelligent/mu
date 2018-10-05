package common

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mholt/archiver"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

// ExtensionImpl provides API for an extension
type ExtensionImpl interface {
	ID() string
	DecorateStackTemplate(assetName string, stackName string, templateBody io.Reader) (io.Reader, error)
	DecorateStackParameters(stackName string, stackParameters map[string]string) (map[string]string, error)
	DecorateStackTags(stackName string, stackTags map[string]string) (map[string]string, error)
}

// BaseExtensionImpl basic no-op extension
type BaseExtensionImpl struct {
	id string
}

// DecorateStackTemplate don't decorate, just return
func (ext *BaseExtensionImpl) DecorateStackTemplate(assetName string, stackName string, inTemplate io.Reader) (io.Reader, error) {
	return inTemplate, nil
}

// DecorateStackParameters don't decorate, just return
func (ext *BaseExtensionImpl) DecorateStackParameters(stackName string, stackParameters map[string]string) (map[string]string, error) {
	return stackParameters, nil
}

// DecorateStackTags don't decorate, just return
func (ext *BaseExtensionImpl) DecorateStackTags(stackName string, stackTags map[string]string) (map[string]string, error) {
	return stackTags, nil
}

// ID returns unique id for extension
func (ext *BaseExtensionImpl) ID() string {
	return ext.id
}

// ExtensionsManager provides API for running extensions
type ExtensionsManager interface {
	ExtensionImpl
	AddExtension(extension ExtensionImpl) error
}

// Implementation of ExtensionsManager
type extensionsManager struct {
	BaseExtensionImpl
	extensions []ExtensionImpl
}

// Create a new extensionsManager
func newExtensionsManager() (ExtensionsManager, error) {
	extMgr := &extensionsManager{
		BaseExtensionImpl{""},
		make([]ExtensionImpl, 0),
	}
	return extMgr, nil
}
func (extMgr *extensionsManager) AddExtension(extension ExtensionImpl) error {
	if extension == nil {
		return fmt.Errorf("extension was nil")
	}
	// ensure extension isn't already loaded
	for _, existingExt := range extMgr.extensions {
		if existingExt.ID() == extension.ID() {
			return fmt.Errorf("extension '%s' already loaded...skipping", extension.ID())
		}
	}
	extMgr.extensions = append(extMgr.extensions, extension)
	return nil
}

// DecorateStackTemplate for all extensions
func (extMgr *extensionsManager) DecorateStackTemplate(assetName string, stackName string, inTemplate io.Reader) (io.Reader, error) {
	outTemplate := inTemplate
	for _, ext := range extMgr.extensions {
		var err error
		outTemplate, err = ext.DecorateStackTemplate(assetName, stackName, outTemplate)
		if err != nil {
			return nil, err
		}
	}
	return outTemplate, nil
}

// DecorateStackParameters for all extensions
func (extMgr *extensionsManager) DecorateStackParameters(stackName string, stackParameters map[string]string) (map[string]string, error) {
	outParams := stackParameters
	for _, ext := range extMgr.extensions {
		var err error
		outParams, err = ext.DecorateStackParameters(stackName, outParams)
		if err != nil {
			return nil, err
		}
	}
	return outParams, nil
}

// DecorateStackTags for all extensions
func (extMgr *extensionsManager) DecorateStackTags(stackName string, stackTags map[string]string) (map[string]string, error) {
	outTags := stackTags
	for _, ext := range extMgr.extensions {
		var err error
		outTags, err = ext.DecorateStackTags(stackName, outTags)
		if err != nil {
			return nil, err
		}
	}
	return outTags, nil
}

// Extension for template overrides in mu.yml
type templateOverrideExtension struct {
	BaseExtensionImpl
	stackNameMatcher *regexp.Regexp
	decoration       interface{}
}

func newTemplateOverrideExtension(stackNamePattern string, template interface{}) ExtensionImpl {
	id := fmt.Sprintf("templateOverride:%s", stackNamePattern)
	ext := &templateOverrideExtension{
		BaseExtensionImpl{id},
		regexp.MustCompile(fmt.Sprintf("^%s$", stackNamePattern)),
		template,
	}
	return ext
}

// DecorateStackTemplate from overrides in mu.yml
func (ext *templateOverrideExtension) DecorateStackTemplate(assetName string, stackName string, inTemplate io.Reader) (io.Reader, error) {
	if stackName != "" && ext.stackNameMatcher.MatchString(stackName) {
		return decorateTemplate(inTemplate, ext.decoration)
	}
	return inTemplate, nil
}

// Extension for tag overrides in mu.yml
type tagOverrideExtension struct {
	BaseExtensionImpl
	stackNameMatcher *regexp.Regexp
	tags             map[string]string
}

func newTagOverrideExtension(stackNamePattern string, tags map[string]string) ExtensionImpl {
	id := fmt.Sprintf("tagOverride:%s", stackNamePattern)
	ext := &tagOverrideExtension{
		BaseExtensionImpl{id},
		regexp.MustCompile(fmt.Sprintf("^%s$", stackNamePattern)),
		tags,
	}
	return ext
}

// DecorateStackTags from overrides in mu.yml
func (ext *tagOverrideExtension) DecorateStackTags(stackName string, stackTags map[string]string) (map[string]string, error) {
	if ext.stackNameMatcher.MatchString(stackName) {
		for k, v := range ext.tags {
			stackTags[k] = v
		}
	}
	return stackTags, nil
}

// Extension for param overrides in mu.yml
type paramOverrideExtension struct {
	BaseExtensionImpl
	stackNameMatcher *regexp.Regexp
	params           map[string]string
}

func newParameterOverrideExtension(stackNamePattern string, params map[string]string) ExtensionImpl {
	id := fmt.Sprintf("paramOverride:%s", stackNamePattern)
	ext := &paramOverrideExtension{
		BaseExtensionImpl{id},
		regexp.MustCompile(fmt.Sprintf("^%s$", stackNamePattern)),
		params,
	}
	return ext
}

// DecorateStackParameters from overrides in mu.yml
func (ext *paramOverrideExtension) DecorateStackParameters(stackName string, stackParams map[string]string) (map[string]string, error) {
	if ext.stackNameMatcher.MatchString(stackName) {
		for k, v := range ext.params {
			stackParams[k] = v
		}
	}
	return stackParams, nil
}

// Extension for archives of templates
type templateArchiveExtension struct {
	BaseExtensionImpl
	path string
	mode TemplateUpdateMode
}

// TemplateUpdateMode of valid template update modes
type TemplateUpdateMode string

// list of update modes
const (
	TemplateUpdateReplace TemplateUpdateMode = "replace"
	TemplateUpdateMerge                      = "merge"
)

func loadExtensionFromArchive(ext *templateArchiveExtension,
	artifactManager ArtifactManager,
	extensionURL *url.URL) error {
	// check for existing etag
	etag := ""
	etagBytes, err := ioutil.ReadFile(filepath.Join(ext.path, ".etag"))
	if err == nil {
		etag = string(etagBytes)
	}

	body, etag, err := artifactManager.GetArtifact(extensionURL.String(), etag)
	if err != nil {
		return err
	}

	if body != nil {
		defer body.Close()

		// empty dir
		os.RemoveAll(ext.path)
		os.MkdirAll(ext.path, 0700)

		// write out archive to dir
		err = extractArchive(ext.path, body)
		if err != nil {
			return err
		}

		// if single directory in extension, assume that's the path
		files, err := ioutil.ReadDir(ext.path)
		if err != nil {
			return err
		}

		originalExtPath := ext.path
		if len(files) == 1 && files[0].IsDir() {
			ext.path = filepath.Join(originalExtPath, files[0].Name())
			log.Debugf("Using directory '%s' for extension '%s'", ext.path, extensionURL)
		}

		// write new etag
		err = ioutil.WriteFile(filepath.Join(originalExtPath, ".etag"), []byte(etag), 0644)
		if err != nil {
			return err
		}
		log.Debugf("Loaded extension from '%s' [id=%s]", extensionURL, ext.id)
	} else {
		log.Debugf("Loaded extension from cache [id=%s]", ext.id)
	}
	return nil
}

func newTemplateArchiveExtension(extensionURL *url.URL, artifactManager ArtifactManager) (ExtensionImpl, error) {
	log.Debugf("Loading extension from '%s'", extensionURL)

	userdir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	extensionsDirectory := filepath.Join(userdir, ".mu", "extensions")

	extID := urlToID(extensionURL)
	ext := &templateArchiveExtension{
		BaseExtensionImpl{extensionURL.String()},
		filepath.Join(extensionsDirectory, extID),
		TemplateUpdateMerge,
	}

	if fi, err := os.Stat(extensionURL.Path); extensionURL.Scheme == "file" && err == nil && fi.IsDir() {
		ext.path = extensionURL.Path
		log.Debugf("Loaded extension from '%s'", extensionURL.Path)
	} else {
		err := loadExtensionFromArchive(ext,
			artifactManager,
			extensionURL)
		if err != nil {
			return nil, err
		}
	}

	// try loading the extension manifest
	extManifest := make(map[interface{}]interface{})
	extManifestFile, err := ioutil.ReadFile(filepath.Join(ext.path, "mu-extension.yml"))
	if err == nil {
		err = yaml.Unmarshal(extManifestFile, extManifest)
		if err != nil {
			log.Debugf("error unmarshalling mu-extension.yml: %s", err)
		} else {
			if v, ok := extManifest["templateUpdateMode"]; ok {
				ext.mode = TemplateUpdateMode(v.(string))
			}
		}
	} else {
		log.Debugf("error reading mu-extension.yml: %s", err)
	}

	// log info about the new extension
	if name, ok := extManifest["name"]; ok {
		if version, ok := extManifest["version"]; ok {
			log.Warningf("Loaded extension %s (version=%v)", name, version)
		} else {
			log.Warningf("Loaded extension %s", name)
		}
	} else {
		log.Warningf("Loaded extension %s", extensionURL)
	}

	return ext, nil
}

// DecorateStackTemplate from template files in archive
func (ext *templateArchiveExtension) DecorateStackTemplate(assetName string, stackName string, inTemplate io.Reader) (io.Reader, error) {
	if assetName == "" {
		return inTemplate, nil
	}
	outTemplate := inTemplate
	assetPath := filepath.Join(ext.path, assetName)

	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		log.Debugf("Path missing, trying without directory: %s", assetPath)
		_, assetName = filepath.Split(assetPath)
		assetPath = filepath.Join(ext.path, assetName)
	}

	if ext.mode == TemplateUpdateReplace {
		f, err := os.Open(assetPath)
		if err != nil {
			log.Debugf("Error trying to replace template '%s': %v", assetName, err)
			return inTemplate, nil
		}
		log.Debugf("Replacing input template '%s'", assetName)
		return f, nil
	}

	yamlFile, err := ioutil.ReadFile(assetPath)

	if err != nil {
		log.Debugf("Unable to find asset '%s' in extension '%s': %s", assetName, ext.id, err)
	} else {
		decoration := make(map[interface{}]interface{})
		err = yaml.Unmarshal(yamlFile, decoration)
		if err != nil {
			log.Warningf("Unable to parse asset '%s' in extension '%s': %s", assetName, ext.id, err)
		} else {
			return decorateTemplate(inTemplate, decoration)
		}
	}
	return outTemplate, nil
}

func urlToID(u *url.URL) string {
	h := sha1.New()
	h.Write([]byte(u.String()))
	return hex.EncodeToString(h.Sum(nil))
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
