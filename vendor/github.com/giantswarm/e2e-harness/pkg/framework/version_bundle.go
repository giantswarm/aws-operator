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

// VBVParams holds information which we can use to query  versionbundle
// version information from the installations repository.
type VBVParams struct {
	// Component is the name of an authority inside a versionbundle IndexRelease.
	// e.g. aws-operator
	Component string
	// Provider is the provider of a versionbundle IndexRelease.
	// This can be aws, azure or kvm.
	Provider string
	// Token is a Github token which is authorized to read from the installations
	// repository.
	Token string
	// VType is the version type of a versionbundle IndexRelease which can be
	// either wip or active.
	VType string
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

	content, err := getContent(params.Provider, params.Token)
	if err != nil {
		return "", microerror.Mask(err)
	}

	output, err := extractReleaseVersion(content, params.VType, params.Component)
	if err != nil {
		return "", microerror.Mask(err)
	}

	logger.Log("level", "debug", "message", fmt.Sprintf("tested version '%s'", params.VType))
	logger.Log("level", "debug", "message", fmt.Sprintf("version bundle version '%s'", output))

	return output, nil
}

func GetAuthorities(params *VBVParams) ([]versionbundle.Authority, error) {
	err := checkType(params.VType)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	logger.Log("level", "debug", "message", fmt.Sprintf("tested version '%s'", params.VType))

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
		return fmt.Errorf("'%s' is not a valid version bundle version type", vType)
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

	if len(indexReleases) < 2 {
		return nil, microerror.Maskf(notFoundError, "release index length must be bigger than 2 but got %d", len(indexReleases))
	}

	// TODO At some point we should get rid of "wip" and "current" wording in the code below.
	//
	//	See https://github.com/giantswarm/giantswarm/issues/3313
	//
	sortedReleases := versionbundle.SortIndexReleasesByVersion(indexReleases)
	sort.Sort(sort.Reverse(sortedReleases))
	switch vType {
	case "latest", "wip":
		return sortedReleases[0].Authorities, nil
	case "previous", "current":
		return sortedReleases[1].Authorities, nil
	}

	return nil, microerror.Maskf(notFoundError, "unknown version type %#q", vType)
}
