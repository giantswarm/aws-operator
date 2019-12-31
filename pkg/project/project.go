package project

var (
<<<<<<< HEAD
	bundleVersion        = "7.1.0"
=======
	bundleVersion        = "8.1.0-dev"
>>>>>>> master
	description          = "The aws-operator handles Kubernetes clusters running on a Kubernetes cluster inside of AWS."
	gitSHA               = "n/a"
	name          string = "aws-operator"
	source        string = "https://github.com/giantswarm/aws-operator"
	version              = "n/a"
)

func BundleVersion() string {
	return bundleVersion
}

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
