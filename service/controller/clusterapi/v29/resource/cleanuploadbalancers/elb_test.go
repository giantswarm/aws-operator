package cleanuploadbalancers

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func Test_splitLoadBalancers(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name              string
		loadBalancerNames []*string
		chunkSize         int
		expectedChunks    [][]*string
	}{
		{
			name:              "case 0: empty lb names returns empty chunks",
			loadBalancerNames: []*string{},
			chunkSize:         20,
			expectedChunks:    [][]*string{},
		},
		{
			name: "case 1: single batch",
			loadBalancerNames: []*string{
				aws.String("lb-1"),
				aws.String("lb-2"),
			},
			chunkSize: 20,
			expectedChunks: [][]*string{{
				aws.String("lb-1"),
				aws.String("lb-2"),
			}},
		},
		{
			name: "case 2: multiple even chunks",
			loadBalancerNames: []*string{
				aws.String("lb-1"),
				aws.String("lb-2"),
				aws.String("lb-3"),
				aws.String("lb-4"),
				aws.String("lb-5"),
				aws.String("lb-6"),
			},
			chunkSize: 2,
			expectedChunks: [][]*string{
				{
					aws.String("lb-1"),
					aws.String("lb-2"),
				},
				{
					aws.String("lb-3"),
					aws.String("lb-4"),
				},
				{
					aws.String("lb-5"),
					aws.String("lb-6"),
				},
			},
		},
		{
			name: "case 3: multiple chunks of different sizes",
			loadBalancerNames: []*string{
				aws.String("lb-1"),
				aws.String("lb-2"),
				aws.String("lb-3"),
				aws.String("lb-4"),
				aws.String("lb-5"),
				aws.String("lb-6"),
				aws.String("lb-7"),
			},
			chunkSize: 3,
			expectedChunks: [][]*string{
				{
					aws.String("lb-1"),
					aws.String("lb-2"),
					aws.String("lb-3"),
				},
				{
					aws.String("lb-4"),
					aws.String("lb-5"),
					aws.String("lb-6"),
				},
				{
					aws.String("lb-7"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := splitLoadBalancers(tc.loadBalancerNames, tc.chunkSize)

			if !reflect.DeepEqual(result, tc.expectedChunks) {
				t.Fatalf("chunks == %#v, want %#v", result, tc.expectedChunks)
			}
		})
	}
}
