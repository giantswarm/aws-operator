package main

import (
	"fmt"
	"log"
	"os"

	"github.com/giantswarm/aws-operator/e2e/tests"
)

func main() {
	// architect requires this
	if len(os.Args) > 1 {
		if os.Args[1] == "version" {
			fmt.Println("0.1.0")
			return
		}

		if os.Args[1] == "--help" {
			fmt.Println("Yet another lookout")
			return
		}
	}

	if err := tests.Run(); err != nil {
		log.Printf("error running tests: %v\n", err)
		os.Exit(1)
	}
}
