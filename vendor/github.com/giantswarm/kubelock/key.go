package kubelock

import (
	"time"
)

func defaultedAcquireOptions(options AcquireOptions) AcquireOptions {
	if options.TTL == 0 {
		options.TTL = DefaultTTL
	}

	return options
}

func defaultedReleaseOptions(options ReleaseOptions) ReleaseOptions {
	return options
}

func isExpired(data lockData) bool {
	return data.CreatedAt.Add(data.TTL).Before(time.Now())
}

func lockAnnotation(name string) string {
	return "kubelock.giantswarm.io/" + name
}
