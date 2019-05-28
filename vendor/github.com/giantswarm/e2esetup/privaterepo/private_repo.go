package privaterepo

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Config struct {
	Owner string
	Repo  string
	Token string
}

type PrivateRepo struct {
	owner string
	repo  string
	token string
}

func New(config Config) (*PrivateRepo, error) {
	if config.Owner == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Owner must not be empty", config)
	}
	if config.Repo == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Repo must not be empty", config)
	}
	if config.Token == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Token must not be empty", config)
	}

	p := &PrivateRepo{
		owner: config.Owner,
		repo:  config.Repo,
		token: config.Token,
	}

	return p, nil
}

func (p *PrivateRepo) Content(ctx context.Context, path string) (string, error) {
	var newClient *github.Client
	{
		c := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: p.token},
		))

		newClient = github.NewClient(c)
	}

	var content string
	{
		in := &github.RepositoryContentGetOptions{}

		out, _, _, err := newClient.Repositories.GetContents(ctx, p.owner, p.repo, path, in)
		if err != nil {
			return "", microerror.Mask(err)
		}

		content, err = out.GetContent()
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	return content, nil
}
