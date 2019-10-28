package ipam

import (
	"context"
)

type TestChecker struct {
	proceed bool
}

func NewTestChecker(proceed bool) *TestChecker {
	a := &TestChecker{
		proceed: proceed,
	}

	return a
}

func (c *TestChecker) Check(ctx context.Context, namespace string, name string) (bool, error) {
	return c.proceed, nil
}
