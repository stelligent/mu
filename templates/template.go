package templates

//go:generate packr

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"text/template"

	"github.com/gobuffalo/packr"
)

// NewTemplate will create a temp file with the template for a CFN stack
func NewTemplate(assetName string, data interface{}) (io.Reader, error) {
	return getAsset(assetName, data)
}

// NewPolicy creates a temp file with a stack policy
func NewPolicy(assetName string) (io.Reader, error) {
	return getAsset(fmt.Sprintf("policies/%s", assetName), nil)
}

// TemplateToString takes an io.Reader and converts it to string
func TemplateToString(reader io.Reader) (string, error) {
	bodyBytes := new(bytes.Buffer)
	_, err := bodyBytes.ReadFrom(reader)
	if err != nil {
		return "", err
	}
	return bodyBytes.String(), err
}

func getAsset(assetName string, data interface{}) (io.Reader, error) {
	box := packr.NewBox("./assets")

	asset, err := box.MustString(assetName)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(assetName).Parse(string(asset[:]))
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	bufWriter := bufio.NewWriter(buf)

	err = tmpl.Execute(bufWriter, data)
	if err != nil {
		return nil, err
	}

	bufWriter.Flush()

	return buf, nil
}
