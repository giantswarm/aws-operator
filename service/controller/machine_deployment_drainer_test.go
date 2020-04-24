package controller

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/giantswarm/microerror"
	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "update .golden CF template file")

// Test_Controller_MachineDeploymentDrainer_Validate tests
func Test_Controller_MachineDeploymentDrainer_Validate(t *testing.T) {
	templateBody, err := Validate("MachineDeploymentDrainer")
	if err != nil {
		t.Fatal(err)
	}

	p := filepath.Join("testdata", "machine_deployment_drainer.golden")

	if *update {
		err := ioutil.WriteFile(p, []byte(templateBody), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	goldenFile, err := ioutil.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal([]byte(templateBody), goldenFile) {
		t.Fatalf("\n\n%s\n", cmp.Diff(string(goldenFile), templateBody))
	}
}

const (
	prefix = "// +operatorkit:validation:controller="
)

func Validate(controller string) (string, error) {
	// Find the caller and its file name. We expect a go test file and derive the
	// source code from it where the controller instantiation resides. Based on
	// the controller file's source code we create an *ast.File which we use for
	// further lookups below.
	var f *ast.File
	{
		_, c, _, _ := runtime.Caller(1)
		t := strings.Replace(c, "_test.go", ".go", 1)

		b, err := ioutil.ReadFile(t)
		if err != nil {
			return "", microerror.Mask(err)
		}

		f, err = parser.ParseFile(token.NewFileSet(), "", string(b), parser.ParseComments)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	// Somewhere in the file an operatorkit controller must be defined and with it
	// the validation tag. In case we do not find it we are pretty much done here.
	var c string
	{
		c = controllerFor(f, prefix)
		if c == "" {
			return "", nil
		}
	}

	var x string
	{
		c = xxx(f, prefix)
		if c == "" {
			return "", nil
		}
		fmt.Printf("%#v\n", x)
	}

	return "", nil
}

func controllerFor(f *ast.File, prefix string) string {
	var c string

	ast.Inspect(f, func(n ast.Node) bool {
		g, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

		if g.Doc == nil {
			return true
		}

		for _, l := range g.Doc.List {
			if strings.HasPrefix(l.Text, prefix) {
				c = strings.Replace(l.Text, prefix, "", 1)
				return false
			}
		}

		return true
	})

	return c
}

func xxx(f *ast.File, prefix string) string {
	var x string

	ast.Inspect(f, func(n ast.Node) bool {
		i, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		fmt.Printf("%#v\n", i.Name)

		// if i.Obj == nil {
		// 	return true
		// }
		//
		// v, ok := i.Obj.Decl.(*ast.ValueSpec)
		// if !ok {
		// 	return true
		// }

		// s, ok := v.Type.(*ast.StarExpr)
		// if !ok {
		// 	return true
		// }

		// e, ok := s.X.(*ast.SelectorExpr)
		// if !ok {
		// 	return true
		// }

		// fmt.Printf("\n")
		// fmt.Printf("\n")
		// fmt.Printf("\n")
		// fmt.Printf("%#v\n", e.X)
		// fmt.Printf("%#v\n", e.Sel)
		// fmt.Printf("\n")
		// fmt.Printf("\n")
		// fmt.Printf("\n")

		return true
	})

	return x
}
