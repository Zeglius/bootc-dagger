package main

import (
	"context"
	"dagger/ci/internal/dagger"
)

// copyCtrMeta copies container metadata (labels and entrypoint) from metaHolder
// to the target container ctr. It returns the modified container and any error encountered.
func copyCtrMeta(ctx context.Context, ctr *dagger.Container, metaHolder *dagger.Container) (*dagger.Container, error) {
	res := ctr

	// Reapply labels to new squashed container
	if labels, err := metaHolder.Labels(ctx); err != nil {
		return nil, err
	} else {
		for _, l := range labels {
			name, _ := l.Name(ctx)
			value, _ := l.Value(ctx)
			res = res.WithLabel(name, value)
		}
	}

	// Reapply entrypoint
	if entrypoint, err := metaHolder.Entrypoint(ctx); err != nil {
		return nil, err
	} else {
		res = res.WithEntrypoint(entrypoint)
	}

	return res, nil
}

// squashCtr creates a new container with a squashed filesystem from the input container.
// It copies over the root filesystem and preserves container metadata (labels and entrypoint).
func squashCtr(ctx context.Context, ctr *dagger.Container) (*dagger.Container, error) {
	squashedCtr := dag.Container().
		WithRootfs(ctr.Rootfs())

	return copyCtrMeta(ctx, squashedCtr, ctr)
}
