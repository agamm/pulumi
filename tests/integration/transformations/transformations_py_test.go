// Copyright 2016-2020, Pulumi Corporation.  All rights reserved.
//go:build (python || all) && !smoke

package ints

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
)

func TestPythonTransformations(t *testing.T) {
	t.Parallel()

	for _, dir := range Dirs {
		d := filepath.Join("python", dir)
		t.Run(d, func(t *testing.T) {
			integration.ProgramTest(t, &integration.ProgramTestOptions{
				Dir: d,
				Dependencies: []string{
					filepath.Join("..", "..", "..", "sdk", "python", "env", "src"),
				},
				Quick:                  true,
				ExtraRuntimeValidation: Validator("python"),
			})
		})
	}
}
