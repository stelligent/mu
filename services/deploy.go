package services

import(
	"fmt"
)

// Deploy a service
func Deploy(environment string, service string) {
	fmt.Printf("deploying service: %s to environment: %s\n",service, environment)
}
