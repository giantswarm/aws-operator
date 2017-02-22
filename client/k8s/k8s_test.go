package k8s

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

func TestGetRawClientConfig(t *testing.T) {
	tests := []struct {
		name                string
		config              Config
		expectedHost        string
		expectedUsername    string
		expectedPassword    string
		expectedBearerToken string
		expectedCertFile    string
		expectedKeyFile     string
		expectedCAFile      string
		expectedCertData    []byte
		expectedKeyData     []byte
		expectedCAData      []byte
	}{
		{
			name: "Specify only in-cluster config. It should return it. Use basic auth.",
			config: Config{
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host:     "http://in-cluster-host",
						Username: "in-cluster-user",
						Password: "in-cluster-password",
					}, nil
				},
			},
			expectedHost:     "http://in-cluster-host",
			expectedUsername: "in-cluster-user",
			expectedPassword: "in-cluster-password",
		},
		{
			name: "Specify only in-cluster config. It should return it. Use token auth.",
			config: Config{
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host:        "http://in-cluster-host",
						BearerToken: "53b3d9d43417971220107dafd2f72def8bd26d2ba93ad4b1eccab98e89b7371f",
					}, nil
				},
			},
			expectedHost:        "http://in-cluster-host",
			expectedBearerToken: "53b3d9d43417971220107dafd2f72def8bd26d2ba93ad4b1eccab98e89b7371f",
		},
		{
			name: "Specify only in-cluster config. It should return it. Use cert auth files.",
			config: Config{
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host: "http://in-cluster-host",
						TLSClientConfig: rest.TLSClientConfig{
							CertFile: "/var/run/kubernetes/client-admin.crt",
							KeyFile:  "/var/run/kubernetes/client-admin.key",
							CAFile:   "/var/run/kubernetes/server-ca.crt",
						},
					}, nil
				},
			},
			expectedHost:     "http://in-cluster-host",
			expectedCertFile: "/var/run/kubernetes/client-admin.crt",
			expectedKeyFile:  "/var/run/kubernetes/client-admin.key",
			expectedCAFile:   "/var/run/kubernetes/server-ca.crt",
		},
		{
			name: "Specify only in-cluster config. It should return it. Use cert auth.",
			config: Config{
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host: "http://in-cluster-host",
						TLSClientConfig: rest.TLSClientConfig{
							CertData: []byte("MIICUjCCAbugAwIBAgIJAPWGX4Ey8qxyMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNV BAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQgQ29tcGFueSBMdGQwHhcNMTcwMjIxMTMwODMxWhcNMTgwMjIxMTMwODMxWjBCMQswCQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNEZWZhdWx0IENvbXBhbnkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDnOUTdUz12uyVnY0H71zEzyrMKyVmdxBWzkKK5RNEVDVCwDSs0Le07V5Y5OAh/2NBZBnt6zwpjBB1Bu6jlZ6bJHCGfnbi9/XWcSjgOLb++IFA63s5G/0t4dyn4vLr8pZawTyen7sWCNDBFYepbaMFMzbkXmaiUDsfWQtpph0uVEwIDAQABo1AwTjAdBgNVHQ4EFgQUrzs82GxwgCNx6gXBsHG6gx2n80cwHwYDVR0jBBgwFoAUrzs82GxwgCNx6gXBsHG6gx2n80cwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQBoo1cQ2wnPkdwwvxvNWn5XEA51iMGD3Gl20xLj54Tm6LEVairRbRG0CzsQalrL2qYCJWRqNWb4t+rZaDpRC/hIJ8fcED+DVjt5gPmj7FLgOTZCFPwe6ERHqT+6L2rtyahT8XG2gXz1LtecrQ9+zCJTwSGOg39I4H+g1Lzq4jwHVQ=="),
							KeyData:  []byte("MIICxjBABgkqhkiG9w0BBQ0wMzAbBgkqhkiG9w0BBQwwDgQIWouPqRB6QdcCAggAMBQGCCqGSIb3DQMHBAhMeW//3DxqkgSCAoBmhJQ0FYLFbxRuVPM3+Dng2r3c4773HTd+a8Yns9vUuJmbpKjI1jXOG3nSLl2SaUSoXXSRP5B0T72ncZ4TezbwgfVBIoEwpfPdzgzkjVM1WKbqTpAuILCyc255XJ3oIl3KxcyTTMCTJBCRrVgj/bDOtLYfiYAwXo5teWJEITJnc9q5Tya+Z1DrClvmAiNax+lKHRHPLyNU1XsfFmIQw7kMsY+VH6z12K0P1Hyt0YUa1FAjhHvNLqP2n2c0Heb6kYa6t3HbN94qkwAcbaZNo3VJxOat7VZmv3ftTep1AgmZLt3ObFo9vANInZINDoa4BUvNwID4hf8QQkKHEPwqCvn4bWjTX2rWEJeJfcTtGoUpzGl4Ptv5/1dwjh6AGzN7OplR1Ho28EjKYES40RWEhRsSkrlSmXmbIFg7bLmMh1exu4JI8H3g8t6B96YSdT0sRuoPO3Yf2w/f2x/RDriWm8Ndc54tpt6fqKdQkTRMvYfTBSkDtJJn6AiB8ApAI5SVmm0Y/dFDP+zWFJz88BNV2oWdD8Tzqcgqn/744YTe8FRErR3m+2vNBGLRvxDG751baAdp+YdUmhGHZwOEzJfk8X+71zl7zGO0Ga93cPcZL0k4zIXuBQKwGePnIHmrkxUo96M/IEIvp3fc5Poi778qpzMhYa7zyzAq6x69r0/U6ux0j4z50MH38HRrH6n5YeRGrWGTxbP1M3EG5aT/O69zHPHZPseDv3eId2LeUpnQO68BLpEpNMineD8XNGHPTHQtbaHYmGkYTGeuIUo0+xTY8qQQQoc2NHSW7HgS0wM3YXZTKiVcXrswHxLmTAdCCoedSubAFa7ut/20WX36iSUM2iDy"),
							CAData:   []byte("MIICXQIBAAKBgQC8xGgc6bpor5MxgH/qfxg6TC6f1+ORyqXKpRCt/UxoS6ko0KOpgk/pYS/kz6GVbb6prqYasrZZQkAWaQNZ4ydvPl17FKNiu5xl2pC8Dq4+dEiuldL4JmQ3RwpVWr9tN6X/wpdQl/5/UfPSqHnKbO2pU2uDWaBeeijWEu7L9S9bYwIDAQABAoGBAJ5k8DfSp+hP62MOQEe0fc/tPPJDZWFged2gxG46rXKWiksFR09lWUirlFSbJSsN+37GXfrpGrmrLbugQn+aa+sq6e2Bo0IZz9ayHC7Lvt3EtzNOnAtKLALDosMoG/tby4m/aaQUZnNkH3/9GeuxO+jb/b7pTJWZb/0kyHJ3UP4BAkEA6nTdO+Jzp3fqWuhUF/hBHAQ7n7jlvTy4ourubE/w9lMH7LueDa2mxg9E2kv15vafUrGXuZ0VEdGi+UVSeCJuMQJBAM4cylWCkuG1Rq5Facn1CG0Zglo19QkJV9Y5cJBK0sTUeB6c5DreB8tW3fcsX8FxJScN9is+W3qjN3j5lcsX2dMCQQDKcgB57g5pY50DxCqgy+cElw8Y2qHdZioT2wHmmpx5RbbJDjPqobAowxRz3jVFqlxmHhzh1CZWTYsI7HfKbghxAkAnFfaYuKY5/zJkIe2pyrnKVqgNi2XoTMlHaqUZ99Z4VQJia8YsE6bOvK5jDRsrh9VPzqn8EVsvqnv+iPYLCX7ZAkA2+AFPirC+Lfrc5fnSDglhvRSNSAqKYHH8VML06lTfHaojb1N6JA2DuAKZwuS8xiBji2FAS2ICFi4PLHe1pQRb"),
						},
					}, nil
				},
			},
			expectedHost:     "http://in-cluster-host",
			expectedCertData: []byte("MIICUjCCAbugAwIBAgIJAPWGX4Ey8qxyMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNV BAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQgQ29tcGFueSBMdGQwHhcNMTcwMjIxMTMwODMxWhcNMTgwMjIxMTMwODMxWjBCMQswCQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNEZWZhdWx0IENvbXBhbnkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDnOUTdUz12uyVnY0H71zEzyrMKyVmdxBWzkKK5RNEVDVCwDSs0Le07V5Y5OAh/2NBZBnt6zwpjBB1Bu6jlZ6bJHCGfnbi9/XWcSjgOLb++IFA63s5G/0t4dyn4vLr8pZawTyen7sWCNDBFYepbaMFMzbkXmaiUDsfWQtpph0uVEwIDAQABo1AwTjAdBgNVHQ4EFgQUrzs82GxwgCNx6gXBsHG6gx2n80cwHwYDVR0jBBgwFoAUrzs82GxwgCNx6gXBsHG6gx2n80cwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQBoo1cQ2wnPkdwwvxvNWn5XEA51iMGD3Gl20xLj54Tm6LEVairRbRG0CzsQalrL2qYCJWRqNWb4t+rZaDpRC/hIJ8fcED+DVjt5gPmj7FLgOTZCFPwe6ERHqT+6L2rtyahT8XG2gXz1LtecrQ9+zCJTwSGOg39I4H+g1Lzq4jwHVQ=="),
			expectedKeyData:  []byte("MIICxjBABgkqhkiG9w0BBQ0wMzAbBgkqhkiG9w0BBQwwDgQIWouPqRB6QdcCAggAMBQGCCqGSIb3DQMHBAhMeW//3DxqkgSCAoBmhJQ0FYLFbxRuVPM3+Dng2r3c4773HTd+a8Yns9vUuJmbpKjI1jXOG3nSLl2SaUSoXXSRP5B0T72ncZ4TezbwgfVBIoEwpfPdzgzkjVM1WKbqTpAuILCyc255XJ3oIl3KxcyTTMCTJBCRrVgj/bDOtLYfiYAwXo5teWJEITJnc9q5Tya+Z1DrClvmAiNax+lKHRHPLyNU1XsfFmIQw7kMsY+VH6z12K0P1Hyt0YUa1FAjhHvNLqP2n2c0Heb6kYa6t3HbN94qkwAcbaZNo3VJxOat7VZmv3ftTep1AgmZLt3ObFo9vANInZINDoa4BUvNwID4hf8QQkKHEPwqCvn4bWjTX2rWEJeJfcTtGoUpzGl4Ptv5/1dwjh6AGzN7OplR1Ho28EjKYES40RWEhRsSkrlSmXmbIFg7bLmMh1exu4JI8H3g8t6B96YSdT0sRuoPO3Yf2w/f2x/RDriWm8Ndc54tpt6fqKdQkTRMvYfTBSkDtJJn6AiB8ApAI5SVmm0Y/dFDP+zWFJz88BNV2oWdD8Tzqcgqn/744YTe8FRErR3m+2vNBGLRvxDG751baAdp+YdUmhGHZwOEzJfk8X+71zl7zGO0Ga93cPcZL0k4zIXuBQKwGePnIHmrkxUo96M/IEIvp3fc5Poi778qpzMhYa7zyzAq6x69r0/U6ux0j4z50MH38HRrH6n5YeRGrWGTxbP1M3EG5aT/O69zHPHZPseDv3eId2LeUpnQO68BLpEpNMineD8XNGHPTHQtbaHYmGkYTGeuIUo0+xTY8qQQQoc2NHSW7HgS0wM3YXZTKiVcXrswHxLmTAdCCoedSubAFa7ut/20WX36iSUM2iDy"),
			expectedCAData:   []byte("MIICXQIBAAKBgQC8xGgc6bpor5MxgH/qfxg6TC6f1+ORyqXKpRCt/UxoS6ko0KOpgk/pYS/kz6GVbb6prqYasrZZQkAWaQNZ4ydvPl17FKNiu5xl2pC8Dq4+dEiuldL4JmQ3RwpVWr9tN6X/wpdQl/5/UfPSqHnKbO2pU2uDWaBeeijWEu7L9S9bYwIDAQABAoGBAJ5k8DfSp+hP62MOQEe0fc/tPPJDZWFged2gxG46rXKWiksFR09lWUirlFSbJSsN+37GXfrpGrmrLbugQn+aa+sq6e2Bo0IZz9ayHC7Lvt3EtzNOnAtKLALDosMoG/tby4m/aaQUZnNkH3/9GeuxO+jb/b7pTJWZb/0kyHJ3UP4BAkEA6nTdO+Jzp3fqWuhUF/hBHAQ7n7jlvTy4ourubE/w9lMH7LueDa2mxg9E2kv15vafUrGXuZ0VEdGi+UVSeCJuMQJBAM4cylWCkuG1Rq5Facn1CG0Zglo19QkJV9Y5cJBK0sTUeB6c5DreB8tW3fcsX8FxJScN9is+W3qjN3j5lcsX2dMCQQDKcgB57g5pY50DxCqgy+cElw8Y2qHdZioT2wHmmpx5RbbJDjPqobAowxRz3jVFqlxmHhzh1CZWTYsI7HfKbghxAkAnFfaYuKY5/zJkIe2pyrnKVqgNi2XoTMlHaqUZ99Z4VQJia8YsE6bOvK5jDRsrh9VPzqn8EVsvqnv+iPYLCX7ZAkA2+AFPirC+Lfrc5fnSDglhvRSNSAqKYHH8VML06lTfHaojb1N6JA2DuAKZwuS8xiBji2FAS2ICFi4PLHe1pQRb"),
		},
		{
			name: "Specify both in-cluster config and CLI config. It should return the in-cluster config. Use basic auth.",
			config: Config{
				Host:     "http://host-from-cli",
				Username: "cli-user",
				Password: "cli-password",
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host:     "http://in-cluster-host",
						Username: "in-cluster-user",
						Password: "in-cluster-password",
					}, nil
				},
			},
			expectedHost:     "http://in-cluster-host",
			expectedUsername: "in-cluster-user",
			expectedPassword: "in-cluster-password",
		},
		{
			name: "Specify both in-cluster config and CLI config. It should return the in-cluster config. Use token auth.",
			config: Config{
				Host:        "http://host-from-cli",
				BearerToken: "20eeff4fea764a6020e767f224ffb2a0ea3fc48bf11e0aadf99c3ee7092e29bd",
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host:        "http://in-cluster-host",
						BearerToken: "53b3d9d43417971220107dafd2f72def8bd26d2ba93ad4b1eccab98e89b7371f",
					}, nil
				},
			},
			expectedHost:        "http://in-cluster-host",
			expectedBearerToken: "53b3d9d43417971220107dafd2f72def8bd26d2ba93ad4b1eccab98e89b7371f",
		},
		{
			name: "Specify both in-cluster config and CLI config. It should return the in-cluster config. Use cert auth files.",
			config: Config{
				Host: "http://host-from-cli",
				TLSClientConfig: TLSClientConfig{
					CertFile: "/var/run/kubernetes/client-cli.crt",
					KeyFile:  "/var/run/kubernetes/client-cli.key",
					CAFile:   "/var/run/kubernetes/server-ca-cli.crt",
				},
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host: "http://in-cluster-host",
						TLSClientConfig: rest.TLSClientConfig{
							CertFile: "/var/run/kubernetes/client-admin.crt",
							KeyFile:  "/var/run/kubernetes/client-admin.key",
							CAFile:   "/var/run/kubernetes/server-ca.crt",
						},
					}, nil
				},
			},
			expectedHost:     "http://in-cluster-host",
			expectedCertFile: "/var/run/kubernetes/client-admin.crt",
			expectedKeyFile:  "/var/run/kubernetes/client-admin.key",
			expectedCAFile:   "/var/run/kubernetes/server-ca.crt",
		},
		{
			name: "Specify both in-cluster config and CLI config. It should return the in-cluster config. Use cert auth.",
			config: Config{
				Host: "http://host-from-cli",
				TLSClientConfig: TLSClientConfig{
					CertData: "MIICUjCCAbugAwIBAgIJAMjeDyoHVLYCMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNVBAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQgQ29tcGFueSBMdGQwHhcNMTcwMjIxMTMyMDIzWhcNMTgwMjIxMTMyMDIzWjBCMQswCQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNEZWZhdWx0IENvbXBhbnkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC06bpDCtf9eZZtI1B4sN1I+10zE33rlQDfO7dIQ3vmXeCUYCQddqjVewlvXsOR+nptG5fjJl/Eac/DRiuucD+gOLUYhyR4YbXkQHOoSkWYmTbrprNT5oAk4hR7auhnkdGC9Dtt1xaLG1s91ciJjlAu4e9gl2lWrfv8GeNKe6QnNwIDAQABo1AwTjAdBgNVHQ4EFgQURIL0u73nFWq4WkmoYsEUkz1QBVcwHwYDVR0jBBgwFoAURIL0u73nFWq4WkmoYsEUkz1QBVcwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQBEKpA4d3o1xHXxA7pXGG3HAHqcBQDRxQvdbBo9FGnei7DpHMUA91lZrd86sZrzl1eL+zWId0B/juhwDTy7JpyvUOxAepctfq9AjZrD8FaDPoZQGQorsHvXq9wvKVo5CsXxq0vJPIJSavbfFV6Fgc2AjJA1tvB5tvm9UJjDRVyUGQ==",
					KeyData:  "MIICxjBABgkqhkiG9w0BBQ0wMzAbBgkqhkiG9w0BBQwwDgQI6qsJivVP2XwCAggAMBQGCCqGSIb3DQMHBAiY3fbmXZHiowSCAoAGP+cA9tc+JnaK5gOZtlmowysq9mmnsQedyU9qgJ9Pl/0nQ+vINUHfrku5Z3yyc2bnrfVKMTGPbvLCkVkmaLJOq2ZEricu6ZVV5OwBbtE1M/aO7ELaSDa57XhJSlfJecbf7nkFlWt5VdSkMMQf99miPeZLLHF44wTkdmcLnyEph/k9DJJ/B6SHWG6MMN7pWm1MWynYegAkleV5Iqf3GawAnqhbswiMRkNVZAzeDQ7R4SAe1XwzT1k/3YWCFU/4J7KKZMlzP0PCg3Rv/dgbZV6RtGx6BwnKf/X+OIkZOHDwZZeeRLTrmZtscavA6DQvnwnwihodKcMebbwxC5DU42ZtqrdYvC7C+5Ih4LDFO//HFNp25ZeFzOzjqSrvI9Y2c88l+1mSfAxMdym9F1zFETlcZngvZzlhWAzqbNhtJbV4cMACt0OyysM4G4c7yvGSDWzqOTrJsgG4GrLdMq3WhKWsWGs6lO3/IdNKv8dkkJkMLay7YgOGjclpnIwYzADIB/ipmKfY8+ClZzVr3U6/3NU/WeLW+zjXHx2p1om6VU9ytzXNwWpY666H/XrENeKbqdHOMAW24Ucm/7+xnyylItyvh+1SexSpwj/PivL2dfwNSph6T+mecJsCs3BzlbNeXsuVTRfMrb2oeRnUpM+nCO/YM6yjKRY1gdvyg2el1twwoDCz1/SpBC7PV/lhfA1rKA+7bttrIyQ6NrxPTI7Kt5Y/behAfZvWRtI5WClG0bfBOEcPXzBGjuFsagFa5Zt7ErClGI5p5C7nsTKIGsUgceyJ5Utk/j9jojCKZ4Yhaa312Yf5NlWHETu0RF4Q3kSEDLH3tIERmXIOgKlomg400GF5",
					CAData:   "MIICXAIBAAKBgQCtH89sRzbTE51DIB/VpVvHspHJB1OmxBaih1nENjuQd205qzK4WtGB6jKqCRfDgjk2u9XKu01CFF548eWCvS2HCKAuNQEvm4MX1fWOxfXVBGlovFtlmi2lSSY6pmJ9Zlgl1SvUPgeR9Je+43Wk4479BJOf++SyD+x4NupvHPa99wIDAQABAoGAJkpV1y39CzxYWQNe5yL2pLlzExJixwyxsOrcyM/x5qbzaoDZ6/pyQhipcgAm2GASBXAP/hHlKYtVxcxCpeLvkYQNfzFXYfEPObNyYtOzevRTkdVzehCOUO/24hlzCpkjYjmGY6DIat1cvr8IPfHYpc652+d7o/WNkGiHFOetkJECQQDVfPdqc+9nBz8WIKeFHgx//xzxS/u65PHcGu4LgI6eBGkH2BX1Gyebq7ts9+kPuK5E1ISbka2eSiW4KTMVoAI/AkEAz5k0X94ZXC7vpOWNleTk3fLVG/F8fJGG97S0dwucy7fn4SZeap7RbCBpnpmfCb8bWR9wIGqNULRI0K1AgE9mSQJAH4rmN2lHvu44KPnMJoPpDuRPj2tNlzCKd53W/AYTjE9UgV8w51UKxhpah+AdJECCJxNLQH0GrPOBnTMhJBnPGwJAX2wgcuB377NzW+xYBEpOGOcBpfJ+MhQCYeGiAgZIcCt8XjVwuLl/sZ/EbK5YN/ar729P7taLVklIHwND3ragYQJBAMlO9huiGxGftFzIy2Aqt/MWJby2L0gdSFeefhJt0pWzSp2IDGONtl5yuZx7vRRk1JFTwKcZGcbown0QkvzUrNU=",
				},
				inClusterConfigProvider: func() (*rest.Config, error) {
					return &rest.Config{
						Host: "http://in-cluster-host",
						TLSClientConfig: rest.TLSClientConfig{
							CertData: []byte("MIICUjCCAbugAwIBAgIJAPWGX4Ey8qxyMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNV BAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQgQ29tcGFueSBMdGQwHhcNMTcwMjIxMTMwODMxWhcNMTgwMjIxMTMwODMxWjBCMQswCQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNEZWZhdWx0IENvbXBhbnkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDnOUTdUz12uyVnY0H71zEzyrMKyVmdxBWzkKK5RNEVDVCwDSs0Le07V5Y5OAh/2NBZBnt6zwpjBB1Bu6jlZ6bJHCGfnbi9/XWcSjgOLb++IFA63s5G/0t4dyn4vLr8pZawTyen7sWCNDBFYepbaMFMzbkXmaiUDsfWQtpph0uVEwIDAQABo1AwTjAdBgNVHQ4EFgQUrzs82GxwgCNx6gXBsHG6gx2n80cwHwYDVR0jBBgwFoAUrzs82GxwgCNx6gXBsHG6gx2n80cwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQBoo1cQ2wnPkdwwvxvNWn5XEA51iMGD3Gl20xLj54Tm6LEVairRbRG0CzsQalrL2qYCJWRqNWb4t+rZaDpRC/hIJ8fcED+DVjt5gPmj7FLgOTZCFPwe6ERHqT+6L2rtyahT8XG2gXz1LtecrQ9+zCJTwSGOg39I4H+g1Lzq4jwHVQ=="),
							KeyData:  []byte("MIICxjBABgkqhkiG9w0BBQ0wMzAbBgkqhkiG9w0BBQwwDgQIWouPqRB6QdcCAggAMBQGCCqGSIb3DQMHBAhMeW//3DxqkgSCAoBmhJQ0FYLFbxRuVPM3+Dng2r3c4773HTd+a8Yns9vUuJmbpKjI1jXOG3nSLl2SaUSoXXSRP5B0T72ncZ4TezbwgfVBIoEwpfPdzgzkjVM1WKbqTpAuILCyc255XJ3oIl3KxcyTTMCTJBCRrVgj/bDOtLYfiYAwXo5teWJEITJnc9q5Tya+Z1DrClvmAiNax+lKHRHPLyNU1XsfFmIQw7kMsY+VH6z12K0P1Hyt0YUa1FAjhHvNLqP2n2c0Heb6kYa6t3HbN94qkwAcbaZNo3VJxOat7VZmv3ftTep1AgmZLt3ObFo9vANInZINDoa4BUvNwID4hf8QQkKHEPwqCvn4bWjTX2rWEJeJfcTtGoUpzGl4Ptv5/1dwjh6AGzN7OplR1Ho28EjKYES40RWEhRsSkrlSmXmbIFg7bLmMh1exu4JI8H3g8t6B96YSdT0sRuoPO3Yf2w/f2x/RDriWm8Ndc54tpt6fqKdQkTRMvYfTBSkDtJJn6AiB8ApAI5SVmm0Y/dFDP+zWFJz88BNV2oWdD8Tzqcgqn/744YTe8FRErR3m+2vNBGLRvxDG751baAdp+YdUmhGHZwOEzJfk8X+71zl7zGO0Ga93cPcZL0k4zIXuBQKwGePnIHmrkxUo96M/IEIvp3fc5Poi778qpzMhYa7zyzAq6x69r0/U6ux0j4z50MH38HRrH6n5YeRGrWGTxbP1M3EG5aT/O69zHPHZPseDv3eId2LeUpnQO68BLpEpNMineD8XNGHPTHQtbaHYmGkYTGeuIUo0+xTY8qQQQoc2NHSW7HgS0wM3YXZTKiVcXrswHxLmTAdCCoedSubAFa7ut/20WX36iSUM2iDy"),
							CAData:   []byte("MIICXQIBAAKBgQC8xGgc6bpor5MxgH/qfxg6TC6f1+ORyqXKpRCt/UxoS6ko0KOpgk/pYS/kz6GVbb6prqYasrZZQkAWaQNZ4ydvPl17FKNiu5xl2pC8Dq4+dEiuldL4JmQ3RwpVWr9tN6X/wpdQl/5/UfPSqHnKbO2pU2uDWaBeeijWEu7L9S9bYwIDAQABAoGBAJ5k8DfSp+hP62MOQEe0fc/tPPJDZWFged2gxG46rXKWiksFR09lWUirlFSbJSsN+37GXfrpGrmrLbugQn+aa+sq6e2Bo0IZz9ayHC7Lvt3EtzNOnAtKLALDosMoG/tby4m/aaQUZnNkH3/9GeuxO+jb/b7pTJWZb/0kyHJ3UP4BAkEA6nTdO+Jzp3fqWuhUF/hBHAQ7n7jlvTy4ourubE/w9lMH7LueDa2mxg9E2kv15vafUrGXuZ0VEdGi+UVSeCJuMQJBAM4cylWCkuG1Rq5Facn1CG0Zglo19QkJV9Y5cJBK0sTUeB6c5DreB8tW3fcsX8FxJScN9is+W3qjN3j5lcsX2dMCQQDKcgB57g5pY50DxCqgy+cElw8Y2qHdZioT2wHmmpx5RbbJDjPqobAowxRz3jVFqlxmHhzh1CZWTYsI7HfKbghxAkAnFfaYuKY5/zJkIe2pyrnKVqgNi2XoTMlHaqUZ99Z4VQJia8YsE6bOvK5jDRsrh9VPzqn8EVsvqnv+iPYLCX7ZAkA2+AFPirC+Lfrc5fnSDglhvRSNSAqKYHH8VML06lTfHaojb1N6JA2DuAKZwuS8xiBji2FAS2ICFi4PLHe1pQRb"),
						},
					}, nil
				},
			},
			expectedHost:     "http://in-cluster-host",
			expectedCertData: []byte("MIICUjCCAbugAwIBAgIJAPWGX4Ey8qxyMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNV BAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQgQ29tcGFueSBMdGQwHhcNMTcwMjIxMTMwODMxWhcNMTgwMjIxMTMwODMxWjBCMQswCQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNEZWZhdWx0IENvbXBhbnkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDnOUTdUz12uyVnY0H71zEzyrMKyVmdxBWzkKK5RNEVDVCwDSs0Le07V5Y5OAh/2NBZBnt6zwpjBB1Bu6jlZ6bJHCGfnbi9/XWcSjgOLb++IFA63s5G/0t4dyn4vLr8pZawTyen7sWCNDBFYepbaMFMzbkXmaiUDsfWQtpph0uVEwIDAQABo1AwTjAdBgNVHQ4EFgQUrzs82GxwgCNx6gXBsHG6gx2n80cwHwYDVR0jBBgwFoAUrzs82GxwgCNx6gXBsHG6gx2n80cwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQBoo1cQ2wnPkdwwvxvNWn5XEA51iMGD3Gl20xLj54Tm6LEVairRbRG0CzsQalrL2qYCJWRqNWb4t+rZaDpRC/hIJ8fcED+DVjt5gPmj7FLgOTZCFPwe6ERHqT+6L2rtyahT8XG2gXz1LtecrQ9+zCJTwSGOg39I4H+g1Lzq4jwHVQ=="),
			expectedKeyData:  []byte("MIICxjBABgkqhkiG9w0BBQ0wMzAbBgkqhkiG9w0BBQwwDgQIWouPqRB6QdcCAggAMBQGCCqGSIb3DQMHBAhMeW//3DxqkgSCAoBmhJQ0FYLFbxRuVPM3+Dng2r3c4773HTd+a8Yns9vUuJmbpKjI1jXOG3nSLl2SaUSoXXSRP5B0T72ncZ4TezbwgfVBIoEwpfPdzgzkjVM1WKbqTpAuILCyc255XJ3oIl3KxcyTTMCTJBCRrVgj/bDOtLYfiYAwXo5teWJEITJnc9q5Tya+Z1DrClvmAiNax+lKHRHPLyNU1XsfFmIQw7kMsY+VH6z12K0P1Hyt0YUa1FAjhHvNLqP2n2c0Heb6kYa6t3HbN94qkwAcbaZNo3VJxOat7VZmv3ftTep1AgmZLt3ObFo9vANInZINDoa4BUvNwID4hf8QQkKHEPwqCvn4bWjTX2rWEJeJfcTtGoUpzGl4Ptv5/1dwjh6AGzN7OplR1Ho28EjKYES40RWEhRsSkrlSmXmbIFg7bLmMh1exu4JI8H3g8t6B96YSdT0sRuoPO3Yf2w/f2x/RDriWm8Ndc54tpt6fqKdQkTRMvYfTBSkDtJJn6AiB8ApAI5SVmm0Y/dFDP+zWFJz88BNV2oWdD8Tzqcgqn/744YTe8FRErR3m+2vNBGLRvxDG751baAdp+YdUmhGHZwOEzJfk8X+71zl7zGO0Ga93cPcZL0k4zIXuBQKwGePnIHmrkxUo96M/IEIvp3fc5Poi778qpzMhYa7zyzAq6x69r0/U6ux0j4z50MH38HRrH6n5YeRGrWGTxbP1M3EG5aT/O69zHPHZPseDv3eId2LeUpnQO68BLpEpNMineD8XNGHPTHQtbaHYmGkYTGeuIUo0+xTY8qQQQoc2NHSW7HgS0wM3YXZTKiVcXrswHxLmTAdCCoedSubAFa7ut/20WX36iSUM2iDy"),
			expectedCAData:   []byte("MIICXQIBAAKBgQC8xGgc6bpor5MxgH/qfxg6TC6f1+ORyqXKpRCt/UxoS6ko0KOpgk/pYS/kz6GVbb6prqYasrZZQkAWaQNZ4ydvPl17FKNiu5xl2pC8Dq4+dEiuldL4JmQ3RwpVWr9tN6X/wpdQl/5/UfPSqHnKbO2pU2uDWaBeeijWEu7L9S9bYwIDAQABAoGBAJ5k8DfSp+hP62MOQEe0fc/tPPJDZWFged2gxG46rXKWiksFR09lWUirlFSbJSsN+37GXfrpGrmrLbugQn+aa+sq6e2Bo0IZz9ayHC7Lvt3EtzNOnAtKLALDosMoG/tby4m/aaQUZnNkH3/9GeuxO+jb/b7pTJWZb/0kyHJ3UP4BAkEA6nTdO+Jzp3fqWuhUF/hBHAQ7n7jlvTy4ourubE/w9lMH7LueDa2mxg9E2kv15vafUrGXuZ0VEdGi+UVSeCJuMQJBAM4cylWCkuG1Rq5Facn1CG0Zglo19QkJV9Y5cJBK0sTUeB6c5DreB8tW3fcsX8FxJScN9is+W3qjN3j5lcsX2dMCQQDKcgB57g5pY50DxCqgy+cElw8Y2qHdZioT2wHmmpx5RbbJDjPqobAowxRz3jVFqlxmHhzh1CZWTYsI7HfKbghxAkAnFfaYuKY5/zJkIe2pyrnKVqgNi2XoTMlHaqUZ99Z4VQJia8YsE6bOvK5jDRsrh9VPzqn8EVsvqnv+iPYLCX7ZAkA2+AFPirC+Lfrc5fnSDglhvRSNSAqKYHH8VML06lTfHaojb1N6JA2DuAKZwuS8xiBji2FAS2ICFi4PLHe1pQRb"),
		},
		{
			name: "Specify only CLI config. It should return it. Use basic auth.",
			config: Config{
				Host:     "http://host-from-cli",
				Username: "cli-user",
				Password: "cli-password",
				inClusterConfigProvider: func() (*rest.Config, error) {
					return nil, fmt.Errorf("No in-cluster config")
				},
			},
			expectedHost:     "http://host-from-cli",
			expectedUsername: "cli-user",
			expectedPassword: "cli-password",
		},
		{
			name: "Specify only CLI config. It should return it. Use token auth.",
			config: Config{
				Host:        "http://host-from-cli",
				BearerToken: "20eeff4fea764a6020e767f224ffb2a0ea3fc48bf11e0aadf99c3ee7092e29bd",
				inClusterConfigProvider: func() (*rest.Config, error) {
					return nil, fmt.Errorf("No in-cluster config")
				},
			},
			expectedHost:        "http://host-from-cli",
			expectedBearerToken: "20eeff4fea764a6020e767f224ffb2a0ea3fc48bf11e0aadf99c3ee7092e29bd",
		},
		{
			name: "Specify only CLI config. It should return it. Use cert auth files.",
			config: Config{
				Host: "http://host-from-cli",
				TLSClientConfig: TLSClientConfig{
					CertFile: "/var/run/kubernetes/client-cli.crt",
					KeyFile:  "/var/run/kubernetes/client-cli.key",
					CAFile:   "/var/run/kubernetes/server-ca-cli.key",
				},
				inClusterConfigProvider: func() (*rest.Config, error) {
					return nil, fmt.Errorf("No in-cluster config")
				},
			},
			expectedHost:     "http://host-from-cli",
			expectedCertFile: "/var/run/kubernetes/client-cli.crt",
			expectedKeyFile:  "/var/run/kubernetes/client-cli.key",
			expectedCAFile:   "/var/run/kubernetes/server-ca-cli.key",
		},
		{
			name: "Specify only CLI config. It should return it. Use cert auth.",
			config: Config{
				Host: "http://host-from-cli",
				TLSClientConfig: TLSClientConfig{
					CertData: "MIICUjCCAbugAwIBAgIJAMjeDyoHVLYCMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNVBAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQgQ29tcGFueSBMdGQwHhcNMTcwMjIxMTMyMDIzWhcNMTgwMjIxMTMyMDIzWjBCMQswCQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNEZWZhdWx0IENvbXBhbnkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC06bpDCtf9eZZtI1B4sN1I+10zE33rlQDfO7dIQ3vmXeCUYCQddqjVewlvXsOR+nptG5fjJl/Eac/DRiuucD+gOLUYhyR4YbXkQHOoSkWYmTbrprNT5oAk4hR7auhnkdGC9Dtt1xaLG1s91ciJjlAu4e9gl2lWrfv8GeNKe6QnNwIDAQABo1AwTjAdBgNVHQ4EFgQURIL0u73nFWq4WkmoYsEUkz1QBVcwHwYDVR0jBBgwFoAURIL0u73nFWq4WkmoYsEUkz1QBVcwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQBEKpA4d3o1xHXxA7pXGG3HAHqcBQDRxQvdbBo9FGnei7DpHMUA91lZrd86sZrzl1eL+zWId0B/juhwDTy7JpyvUOxAepctfq9AjZrD8FaDPoZQGQorsHvXq9wvKVo5CsXxq0vJPIJSavbfFV6Fgc2AjJA1tvB5tvm9UJjDRVyUGQ==",
					KeyData:  "MIICxjBABgkqhkiG9w0BBQ0wMzAbBgkqhkiG9w0BBQwwDgQI6qsJivVP2XwCAggAMBQGCCqGSIb3DQMHBAiY3fbmXZHiowSCAoAGP+cA9tc+JnaK5gOZtlmowysq9mmnsQedyU9qgJ9Pl/0nQ+vINUHfrku5Z3yyc2bnrfVKMTGPbvLCkVkmaLJOq2ZEricu6ZVV5OwBbtE1M/aO7ELaSDa57XhJSlfJecbf7nkFlWt5VdSkMMQf99miPeZLLHF44wTkdmcLnyEph/k9DJJ/B6SHWG6MMN7pWm1MWynYegAkleV5Iqf3GawAnqhbswiMRkNVZAzeDQ7R4SAe1XwzT1k/3YWCFU/4J7KKZMlzP0PCg3Rv/dgbZV6RtGx6BwnKf/X+OIkZOHDwZZeeRLTrmZtscavA6DQvnwnwihodKcMebbwxC5DU42ZtqrdYvC7C+5Ih4LDFO//HFNp25ZeFzOzjqSrvI9Y2c88l+1mSfAxMdym9F1zFETlcZngvZzlhWAzqbNhtJbV4cMACt0OyysM4G4c7yvGSDWzqOTrJsgG4GrLdMq3WhKWsWGs6lO3/IdNKv8dkkJkMLay7YgOGjclpnIwYzADIB/ipmKfY8+ClZzVr3U6/3NU/WeLW+zjXHx2p1om6VU9ytzXNwWpY666H/XrENeKbqdHOMAW24Ucm/7+xnyylItyvh+1SexSpwj/PivL2dfwNSph6T+mecJsCs3BzlbNeXsuVTRfMrb2oeRnUpM+nCO/YM6yjKRY1gdvyg2el1twwoDCz1/SpBC7PV/lhfA1rKA+7bttrIyQ6NrxPTI7Kt5Y/behAfZvWRtI5WClG0bfBOEcPXzBGjuFsagFa5Zt7ErClGI5p5C7nsTKIGsUgceyJ5Utk/j9jojCKZ4Yhaa312Yf5NlWHETu0RF4Q3kSEDLH3tIERmXIOgKlomg400GF5",
					CAData:   "MIICXAIBAAKBgQCtH89sRzbTE51DIB/VpVvHspHJB1OmxBaih1nENjuQd205qzK4WtGB6jKqCRfDgjk2u9XKu01CFF548eWCvS2HCKAuNQEvm4MX1fWOxfXVBGlovFtlmi2lSSY6pmJ9Zlgl1SvUPgeR9Je+43Wk4479BJOf++SyD+x4NupvHPa99wIDAQABAoGAJkpV1y39CzxYWQNe5yL2pLlzExJixwyxsOrcyM/x5qbzaoDZ6/pyQhipcgAm2GASBXAP/hHlKYtVxcxCpeLvkYQNfzFXYfEPObNyYtOzevRTkdVzehCOUO/24hlzCpkjYjmGY6DIat1cvr8IPfHYpc652+d7o/WNkGiHFOetkJECQQDVfPdqc+9nBz8WIKeFHgx//xzxS/u65PHcGu4LgI6eBGkH2BX1Gyebq7ts9+kPuK5E1ISbka2eSiW4KTMVoAI/AkEAz5k0X94ZXC7vpOWNleTk3fLVG/F8fJGG97S0dwucy7fn4SZeap7RbCBpnpmfCb8bWR9wIGqNULRI0K1AgE9mSQJAH4rmN2lHvu44KPnMJoPpDuRPj2tNlzCKd53W/AYTjE9UgV8w51UKxhpah+AdJECCJxNLQH0GrPOBnTMhJBnPGwJAX2wgcuB377NzW+xYBEpOGOcBpfJ+MhQCYeGiAgZIcCt8XjVwuLl/sZ/EbK5YN/ar729P7taLVklIHwND3ragYQJBAMlO9huiGxGftFzIy2Aqt/MWJby2L0gdSFeefhJt0pWzSp2IDGONtl5yuZx7vRRk1JFTwKcZGcbown0QkvzUrNU=",
				},
				inClusterConfigProvider: func() (*rest.Config, error) {
					return nil, fmt.Errorf("No in-cluster config")
				},
			},
			expectedHost:     "http://host-from-cli",
			expectedCertData: []byte("MIICUjCCAbugAwIBAgIJAMjeDyoHVLYCMA0GCSqGSIb3DQEBCwUAMEIxCzAJBgNVBAYTAlhYMRUwEwYDVQQHDAxEZWZhdWx0IENpdHkxHDAaBgNVBAoME0RlZmF1bHQgQ29tcGFueSBMdGQwHhcNMTcwMjIxMTMyMDIzWhcNMTgwMjIxMTMyMDIzWjBCMQswCQYDVQQGEwJYWDEVMBMGA1UEBwwMRGVmYXVsdCBDaXR5MRwwGgYDVQQKDBNEZWZhdWx0IENvbXBhbnkgTHRkMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC06bpDCtf9eZZtI1B4sN1I+10zE33rlQDfO7dIQ3vmXeCUYCQddqjVewlvXsOR+nptG5fjJl/Eac/DRiuucD+gOLUYhyR4YbXkQHOoSkWYmTbrprNT5oAk4hR7auhnkdGC9Dtt1xaLG1s91ciJjlAu4e9gl2lWrfv8GeNKe6QnNwIDAQABo1AwTjAdBgNVHQ4EFgQURIL0u73nFWq4WkmoYsEUkz1QBVcwHwYDVR0jBBgwFoAURIL0u73nFWq4WkmoYsEUkz1QBVcwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOBgQBEKpA4d3o1xHXxA7pXGG3HAHqcBQDRxQvdbBo9FGnei7DpHMUA91lZrd86sZrzl1eL+zWId0B/juhwDTy7JpyvUOxAepctfq9AjZrD8FaDPoZQGQorsHvXq9wvKVo5CsXxq0vJPIJSavbfFV6Fgc2AjJA1tvB5tvm9UJjDRVyUGQ=="),
			expectedKeyData:  []byte("MIICxjBABgkqhkiG9w0BBQ0wMzAbBgkqhkiG9w0BBQwwDgQI6qsJivVP2XwCAggAMBQGCCqGSIb3DQMHBAiY3fbmXZHiowSCAoAGP+cA9tc+JnaK5gOZtlmowysq9mmnsQedyU9qgJ9Pl/0nQ+vINUHfrku5Z3yyc2bnrfVKMTGPbvLCkVkmaLJOq2ZEricu6ZVV5OwBbtE1M/aO7ELaSDa57XhJSlfJecbf7nkFlWt5VdSkMMQf99miPeZLLHF44wTkdmcLnyEph/k9DJJ/B6SHWG6MMN7pWm1MWynYegAkleV5Iqf3GawAnqhbswiMRkNVZAzeDQ7R4SAe1XwzT1k/3YWCFU/4J7KKZMlzP0PCg3Rv/dgbZV6RtGx6BwnKf/X+OIkZOHDwZZeeRLTrmZtscavA6DQvnwnwihodKcMebbwxC5DU42ZtqrdYvC7C+5Ih4LDFO//HFNp25ZeFzOzjqSrvI9Y2c88l+1mSfAxMdym9F1zFETlcZngvZzlhWAzqbNhtJbV4cMACt0OyysM4G4c7yvGSDWzqOTrJsgG4GrLdMq3WhKWsWGs6lO3/IdNKv8dkkJkMLay7YgOGjclpnIwYzADIB/ipmKfY8+ClZzVr3U6/3NU/WeLW+zjXHx2p1om6VU9ytzXNwWpY666H/XrENeKbqdHOMAW24Ucm/7+xnyylItyvh+1SexSpwj/PivL2dfwNSph6T+mecJsCs3BzlbNeXsuVTRfMrb2oeRnUpM+nCO/YM6yjKRY1gdvyg2el1twwoDCz1/SpBC7PV/lhfA1rKA+7bttrIyQ6NrxPTI7Kt5Y/behAfZvWRtI5WClG0bfBOEcPXzBGjuFsagFa5Zt7ErClGI5p5C7nsTKIGsUgceyJ5Utk/j9jojCKZ4Yhaa312Yf5NlWHETu0RF4Q3kSEDLH3tIERmXIOgKlomg400GF5"),
			expectedCAData:   []byte("MIICXAIBAAKBgQCtH89sRzbTE51DIB/VpVvHspHJB1OmxBaih1nENjuQd205qzK4WtGB6jKqCRfDgjk2u9XKu01CFF548eWCvS2HCKAuNQEvm4MX1fWOxfXVBGlovFtlmi2lSSY6pmJ9Zlgl1SvUPgeR9Je+43Wk4479BJOf++SyD+x4NupvHPa99wIDAQABAoGAJkpV1y39CzxYWQNe5yL2pLlzExJixwyxsOrcyM/x5qbzaoDZ6/pyQhipcgAm2GASBXAP/hHlKYtVxcxCpeLvkYQNfzFXYfEPObNyYtOzevRTkdVzehCOUO/24hlzCpkjYjmGY6DIat1cvr8IPfHYpc652+d7o/WNkGiHFOetkJECQQDVfPdqc+9nBz8WIKeFHgx//xzxS/u65PHcGu4LgI6eBGkH2BX1Gyebq7ts9+kPuK5E1ISbka2eSiW4KTMVoAI/AkEAz5k0X94ZXC7vpOWNleTk3fLVG/F8fJGG97S0dwucy7fn4SZeap7RbCBpnpmfCb8bWR9wIGqNULRI0K1AgE9mSQJAH4rmN2lHvu44KPnMJoPpDuRPj2tNlzCKd53W/AYTjE9UgV8w51UKxhpah+AdJECCJxNLQH0GrPOBnTMhJBnPGwJAX2wgcuB377NzW+xYBEpOGOcBpfJ+MhQCYeGiAgZIcCt8XjVwuLl/sZ/EbK5YN/ar729P7taLVklIHwND3ragYQJBAMlO9huiGxGftFzIy2Aqt/MWJby2L0gdSFeefhJt0pWzSp2IDGONtl5yuZx7vRRk1JFTwKcZGcbown0QkvzUrNU="),
		},
	}
	for _, tc := range tests {
		rawClientConfig := getRawClientConfig(tc.config)
		assert.Equal(t, tc.expectedHost, rawClientConfig.Host, fmt.Sprintf("[%s] Hosts should be equal", tc.name))
		assert.Equal(t, tc.expectedUsername, rawClientConfig.Username, fmt.Sprintf("[%s] Usernames should be equal", tc.name))
		assert.Equal(t, tc.expectedPassword, rawClientConfig.Password, fmt.Sprintf("[%s] Passwords should be equal", tc.name))
		assert.Equal(t, tc.expectedBearerToken, rawClientConfig.BearerToken, fmt.Sprintf("[%s] Tokens should be equal", tc.name))
		assert.Equal(t, tc.expectedCertFile, rawClientConfig.TLSClientConfig.CertFile, fmt.Sprintf("[%s] CertFiles should be equal", tc.name))
		assert.Equal(t, tc.expectedKeyFile, rawClientConfig.TLSClientConfig.KeyFile, fmt.Sprintf("[%s] KeyFiles should be equal", tc.name))
		assert.Equal(t, tc.expectedCAFile, rawClientConfig.TLSClientConfig.CAFile, fmt.Sprintf("[%s] CAFiles should be equal", tc.name))
		if tc.expectedCertData != nil {
			assert.Equal(t, tc.expectedCertData, rawClientConfig.TLSClientConfig.CertData, fmt.Sprintf("[%s] CertData should be equal", tc.name))
		}
		if tc.expectedKeyData != nil {
			assert.Equal(t, tc.expectedKeyData, rawClientConfig.TLSClientConfig.KeyData, fmt.Sprintf("[%s] KeyData should be equal", tc.name))
		}
		if tc.expectedCAData != nil {
			assert.Equal(t, tc.expectedCAData, rawClientConfig.TLSClientConfig.CAData, fmt.Sprintf("[%s] CAData should be equal", tc.name))
		}
	}
}
