package environments

import (
	"fmt"
	"text/template"
	"github.com/stelligent/mu/common"
	"github.com/urfave/cli"
	"os"
)
type cfnTemplate struct {
	StackName string
	TemplatePath string
}

func newUpsertCommand(config *common.Config) *cli.Command {
	cmd := &cli.Command{
		Name:      "upsert",
		Aliases:   []string{"up"},
		Usage:     "create/update an environment",
		ArgsUsage: "<environment>",
		Action: func(c *cli.Context) error {
			environmentName := c.Args().First()
			if(len(environmentName) == 0) {
				cli.ShowCommandHelp(c, "upsert")
				return fmt.Errorf("ERROR: environment must be provided!")
			}
			return runUpsert(config, environmentName)
		},
	}

	return cmd
}

func runUpsert(config *common.Config, environmentName string) error {
	// get the environment from config by name
	environment, err := common.GetEnvironment(config, environmentName)

	if err != nil {
		return err
	}

	// generate the CFN template
	template, err := generateCFNTemplate(environment)
	if err != nil {
		return err
	}

	// determine if stack exists

	// create/update the stack

	// wait for stack to be updated

	fmt.Printf("upserting environment:%s stack:%s path:%s\n",environment.Name, template.StackName, template.TemplatePath)

	return nil
}

//go:generate go-bindata -pkg $GOPACKAGE -o assets.go assets/

func generateCFNTemplate(environment *common.Environment) (*cfnTemplate, error) {
	stackName := fmt.Sprintf("mu-env-%s", environment.Name)
	templatePath := fmt.Sprintf("%s%s.yml",os.TempDir(), stackName)

	environmentTemplate, err := Asset("assets/environment-template.yml")
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("environment").Parse(string(environmentTemplate[:]))
	if err != nil {
		return nil, err
	}

	templateOut, err := os.Create(templatePath)
	defer templateOut.Close()
	if err != nil {
		return nil, err
	}

	err = tmpl.Execute(templateOut, environment)
	if err != nil {
		return nil, err
	}

	templateOut.Sync()

	return &cfnTemplate{
		StackName: stackName,
		TemplatePath: templatePath,
	}, nil
}


