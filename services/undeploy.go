package services

import(
	"fmt"
)

func Undeploy(environment string, service string) {
	fmt.Printf("undeploying service: %s to environment: %s\n",service, environment)
}
