package privaterepo

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/valuemodifier/path"
)

func ContentToSSHUserList(content string) (string, error) {
	var err error

	var pathGetter *path.Service
	{
		c := path.Config{
			InputBytes: []byte(content),
			Separator:  ".",
		}

		pathGetter, err = path.New(c)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	v, err := pathGetter.Get("Installation.V1.Guest.SSH.UserList")
	if err != nil {
		return "", microerror.Mask(err)
	}

	return v.(string), nil
}
