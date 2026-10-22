package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func startupWalkSites(name string, src []byte) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), name, src, 0)
	if err != nil {
		return nil, err
	}
	var bad []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if ok && sel.Sel.Name == "Sweep" && fn.Name.Name != "sweepScratch" {
				bad = append(bad, name+":"+fn.Name.Name)
			}
			return true
		})
	}
	return bad, nil
}
func startupWalkGuard(t *testing.T) {
	planted, err := startupWalkSites("second.go", []byte("package main;func secondStartup(){s.Sweep(p,true)}"))
	must(t, err)
	if len(planted) != 1 || planted[0] != "second.go:secondStartup" {
		t.Fatal("planted second walking caller escaped", planted)
	}
	entries, err := os.ReadDir(".")
	must(t, err)
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		src, err := os.ReadFile(entry.Name())
		must(t, err)
		bad, err := startupWalkSites(entry.Name(), src)
		must(t, err)
		if len(bad) > 0 {
			t.Fatalf("recursive scratch walk outside explicit maintenance: %v", bad)
		}
	}
}
