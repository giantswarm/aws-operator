package framework

import (
	"context"
	"fmt"
	"sort"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/versionbundle"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

const (
	defaultOwner = "giantswarm"
	defaultRepo  = "installations"
)

type VBVParams struct {
	Component string
	Provider  string
	Token     string
	VType     string
}

var logger micrologger.Logger

func init() {
	logger, _ = micrologger.New(micrologger.Config{})
}

func GetVersionBundleVersion(params *VBVParams) (string, error) {
	err := checkType(params.VType)
	if err != nil {
		return "", microerror.Mask(err)
	}
	logger.Log("level", "debug", "message", fmt.Sprintf("Tested version %q", params.VType))

	content, err := getContent(params.Provider, params.Token)
	if err != nil {
		return "", microerror.Mask(err)
	}

	output, err := extractReleaseVersion(content, params.VType, params.Component)
	if err != nil {
		return "", microerror.Mask(err)
	}

	logger.Log("level", "debug", fmt.Sprintf("Version Bundle Version %q", output))
	return output, nil
}

func GetAuthorities(params *VBVParams) ([]versionbundle.Authority, error) {
	err := checkType(params.VType)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	logger.Log("level", "debug", "message", fmt.Sprintf("Tested version %q", params.VType))

	content, err := getContent(params.Provider, params.Token)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	authorities, err := extractAuthorities(content, params.VType)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return authorities, nil
}

func getContent(provider, token string) (string, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	path := fmt.Sprintf("release/provider/%s.yaml", provider)
	opt := &github.RepositoryContentGetOptions{}
	repoContent, _, _, err := client.Repositories.GetContents(ctx, defaultOwner, defaultRepo, path, opt)

	if err != nil {
		return "", microerror.Mask(err)
	}

	content, err := repoContent.GetContent()
	if err != nil {
		return "", microerror.Mask(err)
	}

	return content, nil
}

func checkType(vType string) error {
	validVTypes := []string{"", "current", "wip"}
	var isValid bool
	for _, v := range validVTypes {
		if v == vType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("%q is not a valid version bundle version type", vType)
	}

	return nil
}

func extractReleaseVersion(content, vType, component string) (string, error) {
	authorities, err := extractAuthorities(content, vType)
	if err != nil {
		return "", microerror.Mask(err)
	}

	for _, a := range authorities {
		if a.Name == component {
			return a.Version, nil
		}
	}
	return "", microerror.Mask(notFoundError)
}

func extractAuthorities(content, vType string) ([]versionbundle.Authority, error) {
	var indexReleases []versionbundle.IndexRelease

	err := yaml.Unmarshal([]byte(content), &indexReleases)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	sortedReleases := versionbundle.SortIndexReleasesByVersion(indexReleases)
	sort.Sort(sort.Reverse(sortedReleases))
	for _, ir := range sortedReleases {
		if vType == "wip" && !ir.Active || vType == "current" && ir.Active {
			return ir.Authorities, nil
		}
	}
	return nil, microerror.Mask(notFoundError)
}
