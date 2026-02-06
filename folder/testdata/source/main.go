// Package main is an example source file for testing folder workflows.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("Hello from source!")
	oldApi.call("example")
}

func helper() string {
	return "helper function"
}
