package create

import (
	"fmt"
	"testing"

	"github.com/giantswarm/awstpr"
	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
	"github.com/juju/errgo"
	"github.com/stretchr/testify/assert"
)

func TestLoadBalancerName(t *testing.T) {
	tests := []struct {
		desc       string
		domainName string
		tpo        awstpr.CustomObject
		res        string
		err        error
	}{
		{
			desc:       "works",
			domainName: "component.foo.bar.example.com",
			tpo: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "foo-customer",
						},
					},
				},
			},
			res: "foo-customer-component",
		},
		{
			desc:       "also works",
			domainName: "component.of.a.well.formed.domain",
			tpo: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "quux-the-customer",
						},
					},
				},
			},
			res: "quux-the-customer-component",
		},
		{
			desc:       "missing ID key in cloudconfig",
			domainName: "component.foo.bar.example.com",
			tpo: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "",
						},
					},
				},
			},
			res: "",
			err: missingCloudConfigKeyError,
		},
		{
			desc:       "malformed domain name",
			domainName: "not a domain name",
			tpo: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "foo-customer",
						},
					},
				},
			},
			res: "",
			err: malformedCloudConfigKeyError,
		},
		{
			desc:       "missing domain name",
			domainName: "",
			tpo: awstpr.CustomObject{
				Spec: awstpr.Spec{
					Cluster: clustertpr.Spec{
						Cluster: spec.Cluster{
							ID: "foo-customer",
						},
					},
				},
			},
			res: "",
			err: malformedCloudConfigKeyError,
		},
	}

	for _, tc := range tests {
		res, err := loadBalancerName(tc.domainName, tc.tpo)

		if err != nil {
			underlying := errgo.Cause(err)
			assert.Equal(t, tc.err, underlying, fmt.Sprintf("[%s] The input values didn't produce the expected output", tc.desc))
		}

		assert.Equal(t, tc.res, res, fmt.Sprintf("[%s] The input values didn't produce the expected output", tc.desc))
	}
}

func TestComponentName(t *testing.T) {
	tests := []struct {
		desc       string
		domainName string
		res        string
		err        error
	}{
		{
			desc:       "one level of subdomains",
			domainName: "foo.bar.com",
			res:        "foo",
		},
		{
			desc:       "two levels of subdomains",
			domainName: "foo.bar.quux.com",
			res:        "foo",
		},
		{
			desc:       "malformed domain",
			domainName: "not a domain name",
			res:        "",
			err:        malformedCloudConfigKeyError,
		},
		{
			desc:       "empty domain",
			domainName: "",
			res:        "",
			err:        malformedCloudConfigKeyError,
		},
	}

	for _, tc := range tests {
		res, err := componentName(tc.domainName)

		if err != nil {
			assert.True(t, IsMalformedCloudConfigKey(err), fmt.Sprintf("[%s] The input values didn't produce the expected output", tc.desc))
		}

		assert.Equal(t, tc.res, res, fmt.Sprintf("[%s] The input values didn't produce the expected output", tc.desc))
	}
}
