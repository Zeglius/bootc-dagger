package tmpls

import (
	"context"
	"dagger/ci/internal/dagger"
	"maps"
	"text/template"

	"github.com/acobaugh/osrelease"
)

// TmpFuncsWithCtr creates and returns a template.FuncMap containing functions for container operations.
//
// It takes a [context.Context], [dagger.Container] pointer, and [dagger.Client] pointer
// as input and provides access to dagger functions within golang templates.
func TmpFuncsWithCtr(
	ctx context.Context,
	dag *dagger.Client,
) template.FuncMap {
	tmplFuncs := template.FuncMap{}

	maps.Copy(tmplFuncs, template.FuncMap{

		"osRelease": func(address string) (map[string]string, error) {
			// Obtain contents of os-release file
			s, err := dag.Container().
				From(address).
				File("/etc/os-release").
				Contents(ctx)
			if err != nil {
				return nil, err
			}

			return osrelease.ReadString(s)
		},
	})

	return tmplFuncs
}

func fWithCtx[T any](ctx context.Context, f func(context.Context) (T, error)) func() (T, error) {
	return func() (T, error) {
		return f(ctx)
	}
}

var (
	_ any = fWithCtx[any]
)
