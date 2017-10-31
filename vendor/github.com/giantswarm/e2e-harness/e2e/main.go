package main

import (
	"fmt"
	"log"
	"os"

	"github.com/giantswarm/e2e-harness/pkg/results"
	"github.com/spf13/afero"
)

func main() {
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

	ts := &results.TestSuite{
		Tests:    1,
		Failures: 0,
		Errors:   0,
	}
	fs := afero.NewOsFs()
	if err := results.Write(fs, ts); err != nil {
		log.Fatalf("could not write results: %v", err)
	}
}
