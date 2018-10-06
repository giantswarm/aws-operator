package collector

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/prometheus/client_golang/prometheus"

	aws "github.com/aws/aws-sdk-go/aws"
	awsClient "github.com/giantswarm/aws-operator/client/aws"
)

const (
	installationLabel = "installation"
	routeStateLabel   = "state"
)

var (
	routesDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "route_info"),
		"Route information.",
		[]string{
			installationLabel,
			routeStateLabel,
		},
		nil,
	)
)

func (c *Collector) collectAccountsRoutes(ch chan<- prometheus.Metric, clients []awsClient.Clients) {
	var wg sync.WaitGroup

	for _, client := range clients {
		wg.Add(1)
		go func(awsClients awsClient.Clients) {
			defer wg.Done()
			c.collectRoutes(ch, awsClients)
		}(client)
	}

	wg.Wait()
}

func (c *Collector) collectRoutes(ch chan<- prometheus.Metric, awsClients awsClient.Clients) {
	c.logger.Log("level", "debug", "message", "collecting metrics for routes")

	routeNames := []string{
		fmt.Sprintf("%s_private_0", c.installationName),
		fmt.Sprintf("%s_private_1", c.installationName),
	}

	for _, routeName := range routeNames {
		input := &ec2.DescribeRouteTablesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:Name"),
					Values: []*string{&routeName},
				},
			},
		}
		o, err := awsClients.EC2.DescribeRouteTables(input)
		if err != nil {
			c.logger.Log("level", "error", "message", "could not list routes", "stack", fmt.Sprintf("%#v", err))
		}

		for _, rb := range o.RouteTables {
			installationName := installationFromTags(rb.Tags)
			for _, r := range rb.Routes {
				ch <- prometheus.MustNewConstMetric(
					routesDesc,
					prometheus.GaugeValue,
					GaugeValue,
					installationName,
					*r.State,
				)
			}
		}
	}

	c.logger.Log("level", "debug", "message", "finished collecting metrics for routes")
}
