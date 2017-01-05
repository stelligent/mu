package services

import(
	"fmt"
)

// Setenv on a service
func Setenv(environment string, service string, keyvals []string) {
	fmt.Printf("setenv service: %s to environment: %s with vals: %s\n",service, environment, keyvals)
}
