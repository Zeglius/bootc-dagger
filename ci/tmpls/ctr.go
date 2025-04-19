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
			// Create channel for results
			type result struct {
				data map[string]string
				err  error
			}
			ch := make(chan result)

			// Run container operations in goroutine
			go func() {
				s, err := dag.Container().
					From(address).
					File("/etc/os-release").
					Contents(ctx)
				if err != nil {
					ch <- result{nil, err}
					return
				}

				data, err := osrelease.ReadString(s)
				ch <- result{data, err}
			}()

			// Wait for result
			r := <-ch
			return r.data, r.err
		},
	})

	return tmplFuncs
}
