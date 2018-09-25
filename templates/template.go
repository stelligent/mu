package templates

//go:generate packr

import (
	"bufio"
	"bytes"
	"io"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/gobuffalo/packr"
)

// GetAsset reads an asset from "disk" into a string
func GetAsset(assetName string, options ...func(string, *string) error) (string, error) {
	box := packr.NewBox("./assets")

	asset, err := box.MustString(assetName)
	if err != nil {
		return "", err
	}
	assetRef := aws.String(asset)

	for _, option := range options {
		if err := option(assetName, assetRef); err != nil {
			return "", err
		}
	}

	return *assetRef, nil
}

// AddData adds the data parameter to a text/template
func AddData(data interface{}) func(string, *string) error {
	return func(assetName string, asset *string) error {

		tmpl, err := template.New(assetName).Parse(*asset)
		// assetValue := *asset
		// tmpl, err := template.New(assetName).Parse(string(assetValue[:]))
		if err != nil {
			return err
		}

		buf := new(bytes.Buffer)
		bufWriter := bufio.NewWriter(buf)

		err = tmpl.Execute(bufWriter, data)
		if err != nil {
			return err
		}

		bufWriter.Flush()

		templateBodyBytes := new(bytes.Buffer)
		_, err = templateBodyBytes.ReadFrom(buf)
		if err != nil {
			return err
		}
		*asset = templateBodyBytes.String()

		return nil
	}
}

// DecorateTemplate uses an ExtensionImpl to inject data into an extension
func DecorateTemplate(extMgr decoratorImpl,
	stackName string) func(string, *string) error {
	return func(assetName string, asset *string) error {
		assetBuf := bytes.NewBufferString(*asset)
		if _, err := extMgr.DecorateStackTemplate(assetName, stackName, assetBuf); err != nil {
			return err
		}
		templateBodyBytes := new(bytes.Buffer)
		if _, err := templateBodyBytes.ReadFrom(assetBuf); err != nil {
			return err
		}
		*asset = templateBodyBytes.String()
		return nil
	}
}

// stub for decorate template struct to avoid circular dependencies
type decoratorImpl interface {
	DecorateStackTemplate(assetName string, stackName string, templateBody io.Reader) (io.Reader, error)
}
