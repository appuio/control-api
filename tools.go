//go:build tools
// +build tools

// Package tools is a place to put any tooling dependencies as imports.
// Go modules will be forced to download and install them.
package tools

import (
	// This is basically KubeBuilder
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	// To have Kustomize updated via Renovate.
	_ "sigs.k8s.io/kustomize/kustomize/v5"
	// To generate mocks
	_ "go.uber.org/mock/mockgen"
	// To have Kind updated via Renovate
	_ "sigs.k8s.io/kind"
)
