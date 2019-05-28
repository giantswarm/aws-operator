package privaterepo

import (
	"bytes"
	"html/template"
	"net/url"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/valuemodifier/path"
)

func ContentToSSHUserList(content string) (string, error) {
	var err error

	var rendered string
	{
		tmpl, err := template.New("main").Parse(content)
		if err != nil {
			return "", microerror.Mask(err)
		}

		v := struct {
			AWS struct {
				Region      string
				HostCluster struct {
					Account          string
					AdminRoleARN     string
					CloudtrailBucket string
					GuardDuty        bool
				}
				GuestCluster struct {
					Account          string
					AdminRoleARN     string
					CloudtrailBucket string
					GuardDuty        bool
				}
			}
			Active           bool
			Base             string
			Codename         string
			Created          time.Time
			Customer         string
			SolutionEngineer string
			Pipeline         string
			Provider         string
			SSHUser          string
			SSHTunnelNeeded  bool
			Services         map[string]*url.URL
			Jumphosts        map[string]*url.URL
			Machines         map[string]*url.URL
			Updated          time.Time
		}{}

		var w bytes.Buffer
		err = tmpl.Execute(&w, v)
		if err != nil {
			return "", microerror.Mask(err)
		}

		rendered = w.String()
	}

	var pathGetter *path.Service
	{
		c := path.Config{
			InputBytes: []byte(rendered),
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
