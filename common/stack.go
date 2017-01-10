package common

import (
	"fmt"
	"os"
	"text/template"
)

// Stack contains the data about a CloudFormation stack
type Stack struct {
	Name string
	TemplatePath string
}

// NewStack will create a new stack instance
func NewStack(name string) *Stack {
	return &Stack{
		Name: name,
		TemplatePath: fmt.Sprintf("%s%s.yml",os.TempDir(), name),
	}
}

// WriteTemplate will create a temp file with the template for a CFN stack
//go:generate go-bindata -pkg $GOPACKAGE -o assets.go assets/
func (stack *Stack) WriteTemplate(assetName string, data interface{}) (error) {
	asset, err := Asset(fmt.Sprintf("assets/%s",assetName))
	if err != nil {
		return err
	}

	tmpl, err := template.New(assetName).Parse(string(asset[:]))
	if err != nil {
		return err
	}

	templateOut, err := os.Create(stack.TemplatePath)
	defer templateOut.Close()
	if err != nil {
		return err
	}

	err = tmpl.Execute(templateOut, data)
	if err != nil {
		return err
	}

	templateOut.Sync()
	return nil
}
