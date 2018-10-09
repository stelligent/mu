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

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	defPrompt := "no"
	if def {
		defPrompt = "yes"
	}
	answer, err := ui.Ask(message, &input.Options{
		Default:  defPrompt,
		Required: true,
		Loop:     true,
		ValidateFunc: func(s string) error {
			if s != "y" && s != "n" {
				return fmt.Errorf("input must be y or n")
			}
			return nil
		},
	})
	line := strings.ToLower(answer)
	if line == "y" {
		return true, err
	}
	if line == "n" {
		return false, err
	}
	return def, err
}
