package kms

import (
	"fmt"
)

func keyAlias(id string) string {
	return fmt.Sprintf("alias/%s", id)
}
