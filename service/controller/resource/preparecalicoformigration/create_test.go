package preparecalicoformigration

import (
	"fmt"
	"testing"
)

func Test_ensureImageVersionIsUpToDate(t *testing.T) {
	tests := []struct {
		name    string
		image   string
		want    string
		wantErr bool
	}{
		{
			name:    "case 0: image needs replacement",
			image:   "docker.io/giantswarm/node:v3.21.5",
			want:    fmt.Sprintf("docker.io/giantswarm/node:%s", desiredVersion),
			wantErr: false,
		},
		{
			name:    "case 1: image doesn't need replacement",
			image:   "docker.io/giantswarm/node:v3.23.0",
			want:    "docker.io/giantswarm/node:v3.23.0",
			wantErr: false,
		},
		{
			name:    "case 2: image needs replacement in china",
			image:   "fancychinaregistry.cn/giantswarm/node:v3.21.5",
			want:    fmt.Sprintf("fancychinaregistry.cn/giantswarm/node:%s", desiredVersion),
			wantErr: false,
		},
		{
			name:    "case 3: image doesn't need replacement in china",
			image:   "fancychinaregistry.cn/giantswarm/node:v3.23.0",
			want:    "fancychinaregistry.cn/giantswarm/node:v3.23.0",
			wantErr: false,
		},
		{
			name:    "case 4: invalid image",
			image:   "docker.io/giantswarm/node:latest",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ensureImageVersionIsUpToDate(tt.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureImageVersionIsUpToDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ensureImageVersionIsUpToDate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
