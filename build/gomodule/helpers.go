package gomodule

import (
	"github.com/google/blueprint"
)

var (
	// Package context used to define Ninja build rules.
	pctx = blueprint.NewPackageContext("github.com/roman-mazur/bood/gomodule")

	// Ninja rule to execute go build.
	goBuild = pctx.StaticRule("binaryBuild", blueprint.RuleParams{
		Command:     "cd $workDir && go build -o $outputPath $pkg",
		Description: "build go command $pkg",
	}, "workDir", "outputPath", "pkg")

	// Ninja rule to execute go mod vendor.
	goVendor = pctx.StaticRule("vendor", blueprint.RuleParams{
		Command:     "cd $workDir && go mod vendor",
		Description: "vendor dependencies of $name",
	}, "workDir", "name")

	goTest = pctx.StaticRule("gotest", blueprint.RuleParams{
		Command:     "cd $workDir && go test -v $pkg > $outputPath",
		Description: "Build and test $pkg",
	}, "workDir", "pkg", "outputPath")
)

func reportUnresolved(ctx blueprint.ModuleContext, unresolved []string) {
	for _, pattern := range unresolved {
		ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", pattern)
	}
}

type globWithDeps interface {
	GlobWithDeps(src string, exclude []string) ([]string, error)
}

func resolvePatterns(ctx globWithDeps, patterns []string, exclude []string) ([]string, []string) {
	var result = []string{}
	var unresolved = []string{}

	for _, src := range patterns {
		if matches, err := ctx.GlobWithDeps(src, exclude); err == nil {
			result = append(result, matches...)
		} else {
			unresolved = append(unresolved, src)
		}
	}

	return result, unresolved
}
