package common

import (
	"fmt"
	"io"
	"strings"

	"github.com/bobappleyard/readline"
)

// CliPrompt is an object for mocking cli input when testing
type CliPrompt struct{}

// NewCliPrompt returns the pointer to a new CliPrompt struct
func NewCliPrompt() *CliPrompt {
	return new(CliPrompt)
}

// Prompt is the preferred way to prompt the user to answer a yes/no question
func Prompt(message string, def bool) (bool, error) {
	p := NewCliPrompt()
	return p.Prompt(message, def)
}

// String writes a string to the terminal and returns the input
func (cli *CliPrompt) String(prompt string) (string, error) {
	return readline.String(prompt)
}

// Prompt prompts the user to answer a yes/no question
func (cli *CliPrompt) Prompt(message string, def bool) (bool, error) {
	for {
		defPrompt := "y/N"
		if def {
			defPrompt = "Y/n"
		}
		line, err := cli.String(fmt.Sprintf("> %s %s: ", message, defPrompt))
		if err == io.EOF {
			return def, nil
		}
		if err != nil {
			return false, err
		}
		readline.AddHistory(line)
		if line == "" {
			return def, nil
		}
		if strings.ToLower(line) == "y" {
			return true, nil
		}
		if strings.ToLower(line) == "n" {
			return false, nil
		}
	}
}
