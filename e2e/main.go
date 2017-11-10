package main

import (
	"log"
	"os"

	"github.com/giantswarm/aws-operator/e2e/tests"
)

func main() {
	// architect requires this
	if len(os.Args) > 1 {
		return
	}

	if err := tests.Run(); err != nil {
		log.Printf("error running tests: %v\n", err)
		os.Exit(1)
	}
}
