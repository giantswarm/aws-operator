package builder

import "io"

type Builder interface {
	Build(out io.Writer, image, path, tag string, env []string) error
}
