package certs

import "testing"

func Test_NewFilesCluster(t *testing.T) {
	files := NewFilesCluster(Cluster{})

	// Make sure all paths are unique.

	set := map[string]struct{}{}

	for _, f := range files {
		set[f.AbsolutePath] = struct{}{}
	}
	if len(set) != len(files) {
		t.Errorf("expected len = %d, got %d", len(set), len(files))
	}

	// Make sure all paths are absolute.

	for _, f := range files {
		if f.AbsolutePath[0] != '/' {
			t.Errorf("expected to start with %q, got %q", "/", f.AbsolutePath)
		}
	}
}
