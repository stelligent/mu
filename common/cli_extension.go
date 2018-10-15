package common

import (
	"fmt"
	"os"
	"strings"

	input "github.com/tcnksm/go-input"
)

// CliExtension is an interface for defining extended cli actions
type CliExtension interface {
	Prompt(message string, def bool) (bool, error)
}

// CliAdditions exposes methods to prompt the user for cli input
type CliAdditions struct{}

// Prompt prompts the user to answer a yes/no question
func (cli *CliAdditions) Prompt(message string, def bool) (bool, error) {

	ui := NewUI()
	defPrompt := "no"
	if def {
		defPrompt = "yes"
	}
	answer, err := ui.Ask(message, &input.Options{
		Default:  defPrompt,
		Required: true,
		Loop:     true,
		ValidateFunc: func(s string) error {
			if s != "yes" && s != "no" {
				return fmt.Errorf("input must be yes or no")
			}
			return nil
		},
		HideOrder: true,
	})
	line := strings.ToLower(answer)
	if line == "yes" {
		return true, err
	}
	if line == "no" {
		return false, err
	}
	return def, err
}

// GetPasswdPrompt prompts the user to enter a password
func (cli *CliAdditions) GetPasswdPrompt(message string) (string, error) {
	ui := NewUI()
	password, err := ui.Ask(message, &input.Options{
		Required:    true,
		Loop:        true,
		Mask:        true,
		MaskDefault: true,
		HideOrder:   true,
	})
	return strings.TrimSpace(password), err
}

// NewUI returns a new input.UI bound to Std(in/out)
func NewUI() *input.UI {
	return &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
}
