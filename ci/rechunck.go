package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"strings"
)

func (m *Ci) Rechunck(ctx context.Context, ctr *dagger.Container) (*dagger.Container, error) {
	baseRootfs := ctr.Rootfs()

	type PkgName = string
	type PkgFile = string

	// Fetch package names
	pkgs := make(map[PkgName][]PkgFile)
	if s, err := ctr.WithExec([]string{}).Stdout(ctx); err != nil {
		return nil, err
	} else {
		for _, s := range strings.Split(s, "\n") {
			pkgs[s] = nil
		}
	}

	// Fetch package files
	for pkg := range pkgs {
		if s, err := ctr.WithExec([]string{"dnf5", "repoquery", "-q", "--files", pkg}).
			Stdout(ctx); err != nil {
			return nil, err
		} else {
			pkgs[pkg] = strings.Split(s, "\n")
		}
	}

	// TODO(Zeg): Maybe should use diffing to exclue package files from baseRootfs
	// and then reapply them sequentially.

	// Remap into [dagger.Directory], per package
	pkgsFiles := make(map[PkgName]*dagger.Directory)
	for k, v := range pkgs {
		pkgsFiles[k] = baseRootfs.Filter(dagger.DirectoryFilterOpts{Include: v})
	}

	const numLayers = 20
	var layers [numLayers]map[PkgName]*dagger.Directory = [numLayers]map[PkgName]*dagger.Directory{}
	// Initialize layers
	for i := range layers {
		layers[i] = make(map[PkgName]*dagger.Directory)
	}

	// Distribute package files across layers
	layerIndex := 0
	for pkgName, pkgDir := range pkgsFiles {
		layers[layerIndex%numLayers][pkgName] = pkgDir
		layerIndex++
	}

	resRootfs, err := copyCtrMeta(ctx, dag.Container(), ctr)
	if err != nil {
		return nil, err
	}

	// Add package files to result container directory
	for i := range layers {
		for _, dir := range layers[i] {
			if dir != nil {
				resRootfs = resRootfs.WithDirectory("/", dir)
			}
		}
	}

	return ctr, nil
}
