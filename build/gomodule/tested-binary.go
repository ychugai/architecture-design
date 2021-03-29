package gomodule

import (
	"fmt"
	"path"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)
type testedBinaryModule struct {
	blueprint.SimpleName

	properties struct {
		Pkg string // Package name to build
		TestPkg string // Package name to test
		Srcs []string // Source files
		TestSrcs []string // Test source files
		SrcsExclude []string // Exclude patterns
		TestSrcsExclude []string // Test Exclude patterns
		VendorFirst bool // If to call vendor command
		Deps []string // Dependencies
	}
}

func (tb *testedBinaryModule) DynamicDependencies(blueprint.DynamicDependerModuleContext) []string {
	return tb.properties.Deps
}

func (tb *testedBinaryModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding build actions for go binary module '%s'", name)

	outputPath := path.Join(config.BaseOutputDir, "bin", name)
	inputs, unresolved := resolvePatterns(ctx, tb.properties.Srcs, tb.properties.SrcsExclude)
	if len(unresolved) != 0 {
		reportUnresolved(ctx, unresolved)
		return
	}

	testInputs, unresolved := resolvePatterns(ctx, tb.properties.TestSrcs, tb.properties.TestSrcsExclude)
	if len(unresolved) != 0 {
		reportUnresolved(ctx, unresolved)
		return
	}

	if tb.properties.VendorFirst {
		vendorDirPath := path.Join(ctx.ModuleDir(), "vendor")
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Vendor dependencies of %s", name),
			Rule:        goVendor,
			Outputs:     []string{vendorDirPath},
			Implicits:   []string{path.Join(ctx.ModuleDir(), "go.mod")},
			Optional:    true,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"name":    name,
			},
		})
		inputs = append(inputs, vendorDirPath)
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as Go binary", name),
		Rule:        goBuild,
		Outputs:     []string{outputPath},
		Implicits:   inputs,
		Args: map[string]string{
			"outputPath": outputPath,
			"workDir":    ctx.ModuleDir(),
			"pkg":        tb.properties.Pkg,
		},
	})

	testInputs = append(testInputs, outputPath)
	outTestPath := fmt.Sprintf(".%s.test.out", name)
	outTestPath = path.Join(config.BaseOutputDir, outTestPath)
	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Testing %s", name),
		Rule:        goTest,
		Outputs:     []string{outTestPath},
		Implicits:   testInputs,
		Args: map[string]string{
			"outputPath": outTestPath,
			"workDir":    ctx.ModuleDir(),
			"pkg":        tb.properties.TestPkg,
		},
	})
}

func GoBinFactory() (blueprint.Module, []interface{}) {
	mType := &testedBinaryModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
