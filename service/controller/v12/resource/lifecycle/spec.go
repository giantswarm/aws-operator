package lifecycle

import "github.com/aws/aws-sdk-go/service/autoscaling"

type StackState struct {
	Instances     []*autoscaling.Instance
	WorkerASGName string
}
