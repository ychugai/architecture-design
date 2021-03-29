package gomodule

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"github.com/stretchr/testify/assert"
)

type PlugResolver struct {
	errors []string
}

// Checks if slice contains element
func contains(slice []string, el string) bool {
	for _, a := range slice {
		if a == el {
			return true
		}
	}
	return false
}

// Resolves patterns
func (pg *PlugResolver) GlobWithDeps(src string, exclude []string) ([]string, error) {
	// if this patterns should produce error, do it
	if contains(pg.errors, src) {
		return []string{}, fmt.Errorf("Wrong")
	}
	// if this patterns is excluded, return nothing
	if contains(exclude, src) {
		return []string{}, nil
	}
	// Otherwise return this pattern
	return []string{src}, nil
}

func TestResolve(t *testing.T) {
	// Patterns we want to resolve
	patters := []string{"a", "b", "c", "a"}
	// Patterns we don't want to resolve
	exclude := []string{"a"}
	// Patterns that will produce errors
	plug := PlugResolver{errors: []string{"b"}}
	resolved, unresolved := resolvePatterns(&plug, patters, exclude)
	assert.True(t, reflect.DeepEqual(resolved, []string{"c"}))
	assert.True(t, reflect.DeepEqual(unresolved, []string{"b"}))
}

func TestGoBinFactory(t *testing.T) {
	ctx := blueprint.NewContext()

	ctx.MockFileSystem(map[string][]byte{
		"Blueprints": []byte(`
			go_binary {
			  name: "test-out",
			  srcs: ["test-src.go"],
			  pkg: ".",
			  testPkg: ".",
	          vendorFirst: true,
			  srcsExclude: ["**/*_test.go"],
              testSrcs: ["**/*_test.go"],
			}
		`),
		"test-src.go": nil,
	})

	ctx.RegisterModuleType("go_binary", GoBinFactory)

	cfg := bood.NewConfig()

	_, errs := ctx.ParseBlueprintsFiles(".", cfg)
	if len(errs) != 0 {
		t.Fatalf("Syntax errors in the test blueprint file: %s", errs)
	}

	_, errs = ctx.PrepareBuildActions(cfg)
	if len(errs) != 0 {
		t.Errorf("Unexpected errors while preparing build actions: %s", errs)
	}
	buffer := new(bytes.Buffer)
	if err := ctx.WriteBuildFile(buffer); err != nil {
		t.Errorf("Error writing ninja file: %s", err)
	} else {
		text := buffer.String()
		t.Logf("Gennerated ninja build file:\n%s", text)
		if !strings.Contains(text, "out/bin/test-out: ") {
			t.Errorf("Generated ninja file does not have build of the test module")
		}
		if !strings.Contains(text, "test-src.go") {
			t.Errorf("Generated ninja file does not have source dependency")
		}
		if !strings.Contains(text, "build vendor: g.gomodule.vendor | go.mod") {
			t.Errorf("Generated ninja file does not have vendor build rule")
		}
		if !strings.Contains(text, "build out/.test-out.test.out") {
			t.Errorf("Generated ninja file does not have gotest rule")
		}
	}
}
