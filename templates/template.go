package templates

//go:generate packr

import (
	"bufio"
	"bytes"
	"io"
	"text/template"

	"github.com/gobuffalo/packr"
)

// GetAsset reads an asset from "disk" into a string
func GetAsset(assetName string, options ...AssetOption) (string, error) {
	box := packr.NewBox("./assets")

	asset, err := box.MustString(assetName)
	if err != nil {
		return "", err
	}

	for _, option := range options {
		if asset, err = option(assetName, asset); err != nil {
			return "", err
		}
	}

	return asset, nil
}

// ExecuteTemplate executes the data parameter on a text/template
func ExecuteTemplate(data interface{}) AssetOption {
	return func(assetName string, asset string) (string, error) {

		tmpl, err := template.New(assetName).Parse(asset)
		if err != nil {
			return asset, err
		}

		buf := new(bytes.Buffer)
		bufWriter := bufio.NewWriter(buf)

		err = tmpl.Execute(bufWriter, data)
		if err != nil {
			return asset, err
		}

		bufWriter.Flush()

		templateBodyBytes := new(bytes.Buffer)
		_, err = templateBodyBytes.ReadFrom(buf)
		if err != nil {
			return asset, err
		}

		return templateBodyBytes.String(), nil
	}
}

// DecorateTemplate uses an ExtensionImpl to inject data into an extension
func DecorateTemplate(extMgr StackTemplateDecorator,
	stackName string) AssetOption {
	return func(assetName string, asset string) (string, error) {
		assetBuf := bytes.NewBufferString(asset)
		newBuf, err := extMgr.DecorateStackTemplate(assetName, stackName, assetBuf)
		if err != nil {
			return "", err
		}
		templateBodyBytes := new(bytes.Buffer)
		if _, err := templateBodyBytes.ReadFrom(newBuf); err != nil {
			return "", err
		}
		return templateBodyBytes.String(), nil
	}
}

// AssetOption describes the method signature for manipulating loaded assets
type AssetOption func(string, string) (string, error)

// StackTemplateDecorator is a stub for decorate template struct to avoid circular dependencies
type StackTemplateDecorator interface {
	DecorateStackTemplate(assetName string, stackName string, templateBody io.Reader) (io.Reader, error)
}
