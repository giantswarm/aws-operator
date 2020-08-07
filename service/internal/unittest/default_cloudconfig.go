package unittest

import "context"

type CloudConfig struct {
}

func (c *CloudConfig) NewHashes(ctx context.Context, obj interface{}) ([]string, error) {
	return []string{"sha512-foobar1", "sha512-foobar2", "sha512-foobar3"}, nil
}

func (c *CloudConfig) NewPaths(ctx context.Context, obj interface{}) ([]string, error) {
	return nil, nil
}

func (c *CloudConfig) NewTemplates(ctx context.Context, obj interface{}) ([]string, error) {
	return nil, nil
}
