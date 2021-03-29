package gomodule

import (
	"fmt"
	"path"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
)

var (
	goDocs = pctx.StaticRule("godoc", blueprint.RuleParams{
		Command: "cd $workDir && go doc -all -u $name > $outputPath",
		Description: "Generate documentation for $name package",
	}, "workDir", "name", "outputPath")
)

type godocsModule struct {
	blueprint.SimpleName
	properties struct {
		Name string
		Pkg string
		Srcs []string
	}
}

func (gd *godocsModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	
	outputPath := path.Join(config.BaseOutputDir, "docs", fmt.Sprintf("%s.html", name))
	
	inputs, unresolved := resolvePatterns(ctx, gd.properties.Srcs, nil)
	if len(unresolved) != 0 {
		reportUnresolved(ctx, unresolved)
		return
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Generating docs for %s package", name),
		Rule: 		goDocs,
		Outputs: []string{outputPath},
		Implicits: inputs,
		Args: map[string]string{
			"workDir": ctx.ModuleDir(),
			"name": name,
			"outputPath": outputPath,
		},
	})
}

func GodocsFactory() (blueprint.Module, []interface{}) {
	mType := &godocsModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}