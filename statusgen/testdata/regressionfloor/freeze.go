package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
)

func main() {
	p := "testdata/regressionfloor/inventory.yaml"
	b, e := os.ReadFile(p)
	must(e)
	var inv map[string]any
	must(yaml.Unmarshal(b, &inv))
	for _, v := range inv["entries"].([]any) {
		row := v.(map[string]any)
		b, e := exec.Command("git", "--no-replace-objects", "show", row["source"].(string)+":"+row["file"].(string)).Output()
		must(e)
		f, e := parser.ParseFile(token.NewFileSet(), "baseline.go", b, 0)
		must(e)
		found := false
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == row["test"] {
				var out bytes.Buffer
				must(format.Node(&out, token.NewFileSet(), fn.Body))
				row["body_sha256"] = fmt.Sprintf("%x", sha256.Sum256(out.Bytes()))
				found = true
			}
		}
		if !found {
			panic("missing pinned source test")
		}
	}
	b, e = json.MarshalIndent(inv, "", "  ")
	must(e)
	must(os.WriteFile(p, append(b, '\n'), 0644))
}
func must(e error) {
	if e != nil {
		panic(e)
	}
}
