package tccp

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net"
	"path/filepath"
	"strconv"
	"testing"
	"unicode"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/annotation"
	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/detection"
)

const (
	TestClusterID           = "8y5ck"
	TestMachineDeploymentID = "al9qy"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_newTemplateBody tests tenant cluster CloudFormation template rendering.
// It is meant to be used as a tool to easily check resulting CF template and
// prevent from accidental CF template changes.
//
// It uses golden file as reference template and when changes to template are
// intentional, they can be updated by providing -update flag for go test.
//
//  go test ./service/controller/clusterapi/v29/resource/tccp -run Test_newTemplateBody -update
//
func Test_newTemplateBody(t *testing.T) {
	testCases := []struct {
		name         string
		cr           v1alpha1.Cluster
		ctx          controllercontext.Context
		tp           templateParams
		errorMatcher func(error) bool
	}{
		{
			name: "case 0: basic test",
			cr:   defaultCluster(),
			ctx:  defaultControllerContext(),
			tp: templateParams{
				DockerVolumeResourceName:   "rsc-abbacd01",
				MasterInstanceResourceName: "rsc-ac0dc01",
			},
			errorMatcher: nil,
		},
	}

	var r *Resource
	{
		d, err := detection.New(detection.Config{Logger: microloggertest.New()})
		if err != nil {
			t.Fatal(err)
		}

		c := Config{
			Detection:        d,
			Logger:           microloggertest.New(),
			EncrypterBackend: "kms",
			VPCPeerID:        "vpc-f8d0e10b",
		}

		r, err = New(c)
		if err != nil {
			t.Fatal(err)
		}
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx := controllercontext.NewContext(context.Background(), tc.ctx)
			actual, err := r.newTemplateBody(ctx, tc.cr, tc.tp)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			golden := filepath.Join("testdata", normalizeToFileName(tc.name)+".golden")
			if *update {
				ioutil.WriteFile(golden, []byte(actual), 0644)
			}

			expected, err := ioutil.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal([]byte(actual), expected) {
				t.Fatalf("\n\n%s\n", cmp.Diff(actual, string(expected)))
			}
		})
	}
}

func defaultCluster() v1alpha1.Cluster {
	g8sSpec := g8sv1alpha1.AWSClusterSpec{
		Cluster: g8sv1alpha1.AWSClusterSpecCluster{
			Description: "Test cluster for template rendering unit test.",
			DNS: g8sv1alpha1.AWSClusterSpecClusterDNS{
				Domain: "guux.eu-central-1.aws.gigantic.io",
			},
		},
		Provider: g8sv1alpha1.AWSClusterSpecProvider{
			CredentialSecret: g8sv1alpha1.AWSClusterSpecProviderCredentialSecret{
				Name:      "default-credential-secret",
				Namespace: "default",
			},
			Master: g8sv1alpha1.AWSClusterSpecProviderMaster{
				AvailabilityZone: "eu-central-1b",
				InstanceType:     "m5.xlarge",
			},
			Region: "eu-central-1",
		},
	}

	cr := v1alpha1.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster: TestClusterID,
			},
		},
	}

	return withG8sClusterSpec(cr, g8sSpec)
}

func defaultControllerContext() controllercontext.Context {
	return controllercontext.Context{
		Status: controllercontext.ContextStatus{
			ControlPlane: controllercontext.ContextStatusControlPlane{
				AWSAccountID: "control-plane-account",
				NATGateway:   controllercontext.ContextStatusControlPlaneNATGateway{},
				RouteTable:   controllercontext.ContextStatusControlPlaneRouteTable{},
				PeerRole: controllercontext.ContextStatusControlPlanePeerRole{
					ARN: "imaginary-cp-peer-role-arn",
				},
				VPC: controllercontext.ContextStatusControlPlaneVPC{
					CIDR: "10.1.0.0/16",
				},
			},
			TenantCluster: controllercontext.ContextStatusTenantCluster{
				AWS: controllercontext.ContextStatusTenantClusterAWS{
					AccountID: "tenant-account",
				},
				Encryption:            controllercontext.ContextStatusTenantClusterEncryption{},
				HostedZoneNameServers: "1.1.1.1,8.8.8.8",
				MasterInstance:        controllercontext.ContextStatusTenantClusterMasterInstance{},
				TCCP: controllercontext.ContextStatusTenantClusterTCCP{
					ASG: controllercontext.ContextStatusTenantClusterTCCPASG{},
					AvailabilityZones: []controllercontext.ContextTenantClusterAvailabilityZone{
						{
							Name:          "eu-central-1a",
							PrivateSubnet: mustParseCIDR("10.100.3.0/27"),
							PublicSubnet:  mustParseCIDR("10.100.3.32/27"),
						},
						{
							Name:          "eu-central-1b",
							PrivateSubnet: mustParseCIDR("10.100.3.64/27"),
							PublicSubnet:  mustParseCIDR("10.100.3.96/27"),
						},
						{
							Name:          "eu-central-1c",
							PrivateSubnet: mustParseCIDR("10.100.3.128/27"),
							PublicSubnet:  mustParseCIDR("10.100.3.164/27"),
						},
					},
					IsTransitioning:   false,
					MachineDeployment: defaultMachineDeployment(),
					VPC: controllercontext.ContextStatusTenantClusterTCCPVPC{
						PeeringConnectionID: "imagenary-peering-connection-id",
					},
				},
				VersionBundleVersion: "6.3.0",
				WorkerInstance: controllercontext.ContextStatusTenantClusterWorkerInstance{
					DockerVolumeSizeGB: "100",
					Image:              "ami-0eb0d9bb7ad1bd1e9",
					Type:               "m5.xlarge",
				},
			},
		},
		Spec: controllercontext.ContextSpec{
			TenantCluster: controllercontext.ContextSpecTenantCluster{
				TCCP: controllercontext.ContextSpecTenantClusterTCCP{
					AvailabilityZones: []controllercontext.ContextTenantClusterAvailabilityZone{
						{
							Name:          "eu-central-1a",
							PrivateSubnet: mustParseCIDR("10.100.3.0/27"),
							PublicSubnet:  mustParseCIDR("10.100.3.32/27"),
						},
						{
							Name:          "eu-central-1b",
							PrivateSubnet: mustParseCIDR("10.100.3.64/27"),
							PublicSubnet:  mustParseCIDR("10.100.3.96/27"),
						},
						{
							Name:          "eu-central-1c",
							PrivateSubnet: mustParseCIDR("10.100.3.128/27"),
							PublicSubnet:  mustParseCIDR("10.100.3.164/27"),
						},
					},
				},
			},
		},
	}
}

func defaultMachineDeployment() v1alpha1.MachineDeployment {
	g8sSpec := g8sv1alpha1.AWSMachineDeploymentSpec{
		NodePool: g8sv1alpha1.AWSMachineDeploymentSpecNodePool{
			Description: "Test node pool for cluster in template rendering unit test.",
			Machine: g8sv1alpha1.AWSMachineDeploymentSpecNodePoolMachine{
				DockerVolumeSizeGB:  100,
				KubeletVolumeSizeGB: 100,
			},
			Scaling: g8sv1alpha1.AWSMachineDeploymentSpecNodePoolScaling{
				Max: 5,
				Min: 3,
			},
		},
		Provider: g8sv1alpha1.AWSMachineDeploymentSpecProvider{
			AvailabilityZones: []string{
				"eu-central-1a",
				"eu-central-1c",
			},
			Worker: g8sv1alpha1.AWSMachineDeploymentSpecProviderWorker{
				InstanceType: "m5.2xlarge",
			},
		},
	}

	cr := v1alpha1.MachineDeployment{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				annotation.MachineDeploymentSubnet: "10.100.8.0/24",
			},
			Labels: map[string]string{
				label.Cluster:           TestClusterID,
				label.MachineDeployment: TestMachineDeploymentID,
			},
		},
	}

	return withG8sMachineDeploymentSpec(cr, g8sSpec)
}

func mustParseCIDR(s string) net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return *n
}

// normalizeToFileName converts all non-digit, non-letter runes in input string
// to dash ('-'). Coalesces multiple dashes into one.
func normalizeToFileName(s string) string {
	var result []rune
	for _, r := range []rune(s) {
		if unicode.IsDigit(r) || unicode.IsLetter(r) {
			result = append(result, r)
		} else {
			l := len(result)
			if l > 0 && result[l-1] != '-' {
				result = append(result, rune('-'))
			}
		}
	}
	return string(result)
}

func withG8sClusterSpec(cr v1alpha1.Cluster, providerExtension g8sv1alpha1.AWSClusterSpec) v1alpha1.Cluster {
	var err error

	if cr.Spec.ProviderSpec.Value == nil {
		cr.Spec.ProviderSpec.Value = &runtime.RawExtension{}
	}

	cr.Spec.ProviderSpec.Value.Raw, err = json.Marshal(&providerExtension)
	if err != nil {
		panic(err)
	}

	return cr
}

func withG8sMachineDeploymentSpec(cr v1alpha1.MachineDeployment, providerExtension g8sv1alpha1.AWSMachineDeploymentSpec) v1alpha1.MachineDeployment {
	var err error

	if cr.Spec.Template.Spec.ProviderSpec.Value == nil {
		cr.Spec.Template.Spec.ProviderSpec.Value = &runtime.RawExtension{}
	}

	cr.Spec.Template.Spec.ProviderSpec.Value.Raw, err = json.Marshal(&providerExtension)
	if err != nil {
		panic(err)
	}

	return cr
}
