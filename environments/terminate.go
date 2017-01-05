package environments

import(
	"fmt"
)

// Terminate an environment
func Terminate(environment string) {
	fmt.Printf("terminating environment: %s\n",environment)
}
