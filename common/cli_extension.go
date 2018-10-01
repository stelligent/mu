package common

import (
	"fmt"
	"io"
	"strings"

	"github.com/bobappleyard/readline"
)

// CliExtension is an interface for defining extended cli actions
type CliExtension interface {
	Prompt(message string, def bool) (bool, error)
}

// CliAdditions exposes methods to prompt the user for cli input
type CliAdditions struct{}

// String writes a string to the terminal and returns the input
func (cli *CliAdditions) String(prompt string) (string, error) {
	return readline.String(prompt)
}

// Prompt prompts the user to answer a yes/no question
func (cli *CliAdditions) Prompt(message string, def bool) (bool, error) {
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
