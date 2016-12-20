package services

import(
	"fmt"
)

func Deploy(environment string, service string) {
	fmt.Printf("deploying service: %s to environment: %s\n",service, environment)
}
