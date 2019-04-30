package key

import (
	"k8s.io/apimachinery/pkg/api/meta"
)

func IsDeleted(v interface{}) bool {
	m, err := meta.Accessor(v)
	if err != nil {
		panic(err)
	}

	return m.GetDeletionTimestamp() != nil
}
