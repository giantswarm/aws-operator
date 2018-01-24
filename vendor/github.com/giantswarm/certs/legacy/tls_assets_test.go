package legacy

import (
	"reflect"
	"testing"
)

func TestValidComponent(t *testing.T) {
	tests := []struct {
		name       string
		components []ClusterComponent
		el         ClusterComponent
		res        bool
	}{
		{
			name: "el is present",
			components: []ClusterComponent{
				ClusterComponent("foobar"),
			},
			el:  ClusterComponent("foobar"),
			res: true,
		},
		{
			name: "el is not present",
			components: []ClusterComponent{
				ClusterComponent("foo"),
			},
			el:  ClusterComponent("bar"),
			res: false,
		},
		{
			name:       "components is empty",
			components: []ClusterComponent{},
			el:         ClusterComponent("foobar"),
			res:        false,
		},
	}

	for i, tc := range tests {
		res := ValidComponent(tc.el, tc.components)

		if !reflect.DeepEqual(tc.res, res) {
			t.Errorf("case %d: want valid = %v, got %v", i, tc.res, res)
		}
	}
}
