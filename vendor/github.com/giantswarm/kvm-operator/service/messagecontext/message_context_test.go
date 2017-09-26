package messagecontext

import (
	"context"
	"testing"
)

func Test_MessageContext(t *testing.T) {
	ctx := context.Background()

	_, ok := FromContext(ctx)
	if ok {
		t.Fatalf("expected %#v got %#v", false, true)
	}

	ctx = NewContext(ctx, NewMessage())

	m1, ok := FromContext(ctx)
	if !ok {
		t.Fatalf("expected %#v got %#v", true, false)
	}
	l1 := len(m1.ConfigMapNames)
	if l1 != 0 {
		t.Fatalf("expected %#v got %#v", 0, l1)
	}

	m1.ConfigMapNames = append(m1.ConfigMapNames, "config-map-1")

	m2, ok := FromContext(ctx)
	if !ok {
		t.Fatalf("expected %#v got %#v", true, false)
	}
	l2 := len(m2.ConfigMapNames)
	if l2 != 1 {
		t.Fatalf("expected %#v got %#v", 1, l2)
	}
}
