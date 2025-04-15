package main

import (
	"context"
	"dagger/ci/internal/dagger"
)

// Install a starting github workflow in your current working
// directory to build Bootc images with bootc-dagger/ci.
//
// Example:
//
//	dagger -m ${GIT_CI_MODULE?} call install-gh-workflow export --path=.
func (m *Ci) InstallGhWorkflow(ctx context.Context) *dagger.Directory {
	return dag.CurrentModule().Source().
		Directory("install/gha")
}
