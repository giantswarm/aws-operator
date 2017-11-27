package tasks

import "github.com/giantswarm/microerror"

// Task represent a generic step in a pipeline.
type Task func() error

func Run(tasks []Task) error {
	var err error
	for _, task := range tasks {
		err = task()
		if err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}

func RunIgnoreError(tasks []Task) {
	for _, task := range tasks {
		task()
	}
}
