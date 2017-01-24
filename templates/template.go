package templates

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"text/template"
)

// NewTemplate will create a temp file with the template for a CFN stack
//go:generate go-bindata -pkg $GOPACKAGE -o assets.go assets/
func NewTemplate(assetName string, data interface{}) (io.Reader, error) {
	asset, err := Asset(fmt.Sprintf("assets/%s", assetName))
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
