package project

var (
	description        = "The aws-operator manages Kubernetes clusters running on AWS."
	gitSHA             = "n/a"
	name        string = "aws-operator"
	source      string = "https://github.com/giantswarm/aws-operator"
	version            = "8.6.1-hama"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
