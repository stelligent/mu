package templates

//go:generate packr

import (
	"bufio"
	"bytes"
	"io"
	"text/template"

	"github.com/gobuffalo/packr"
)

// NewTemplate will create a temp file with the template for a CFN stack
func NewTemplate(assetName string, data interface{}) (io.Reader, error) {
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
