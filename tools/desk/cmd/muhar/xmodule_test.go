package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// twoModules lays out two sibling Go modules, app (the run's own) and lib (a
// different one), each with one package file, and returns their directories.
func twoModules(t *testing.T, libSrc string) (app, lib string) {
	t.Helper()
	top := t.TempDir()
	app = filepath.Join(top, "app")
	lib = filepath.Join(top, "lib")
	files := map[string]string{
		filepath.Join(app, "go.mod"):          "module example.test/app\n\ngo 1.21\n",
		filepath.Join(app, "a.go"):            "package app\n\nfunc A() int { return 1 }\n",
		filepath.Join(app, "sub", "s.go"):     "package sub\n\nfunc S() int { return 2 }\n",
		filepath.Join(lib, "go.mod"):          "module example.test/lib\n\ngo 1.21\n",
		filepath.Join(lib, "l.go"):            libSrc,
		filepath.Join(lib, "inner", "doc.go"): "package inner\n",
	}
	for p, c := range files {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return app, lib
}

// regression: F-muhar-xmodule-build
func TestBuildMutation_CrossModuleTarget(t *testing.T) {
	app, _ := twoModules(t, "package lib\n\nfunc L() int { return 3 }\n")
	env := os.Environ()
	// A spec run from app mutating a file of the sibling module lib: the build
	// must run in lib, not hand `go` a package outside app's main module.
	m := Mutation{Name: "x", File: "../lib/l.go"}
	res := buildMutation(app, m, "", "go test ./...", env, 0)
	if res.Failed || res.TimedOut || res.Err != nil {
		t.Fatalf("a cross-module mutant that compiles was not built from its own module: %+v", res)
	}
}

// regression: F-muhar-xmodule-build
func TestBuildMutation_CrossModuleBrokenMutantFails(t *testing.T) {
	// The cross-module build must still judge the mutant: a file that does not
	// compile is reported Failed, never waved through as built.
	app, _ := twoModules(t, "package lib\n\nfunc L() int { return \"x\" }\n")
	m := Mutation{Name: "x", File: "../lib/l.go"}
	if res := buildMutation(app, m, "", "go test ./...", os.Environ(), 0); !res.Failed {
		t.Fatalf("a cross-module mutant that does not compile was not reported as unbuildable: %+v", res)
	}
}

// regression: F-muhar-xmodule-build
func TestBuildMutation_CrossModuleSubpackageAndEmptyRoot(t *testing.T) {
	app, lib := twoModules(t, "package lib\n\nfunc L() int { return 3 }\n")
	// An absolute file path in a nested package of the other module.
	m := Mutation{Name: "x", File: filepath.Join(lib, "inner", "doc.go")}
	if res := buildMutation(app, m, "", "go test ./...", os.Environ(), 0); res.Failed || res.Err != nil {
		t.Fatalf("nested cross-module package: %+v", res)
	}
	// An empty root means the process cwd, the in-place sequential default.
	t.Chdir(app)
	m = Mutation{Name: "x", File: "../lib/l.go"}
	if res := buildMutation("", m, "", "go test ./...", os.Environ(), 0); res.Failed || res.Err != nil {
		t.Fatalf("cross-module mutant with an empty root: %+v", res)
	}
}

func TestBuildTarget_SameModuleUnchanged(t *testing.T) {
	app, _ := twoModules(t, "package lib\n")
	for _, tc := range []struct{ file, wantPkg string }{
		{filepath.Join(app, "a.go"), "."},
		{filepath.Join(app, "sub", "s.go"), "./sub"},
	} {
		dir, pkg, err := buildTarget(app, tc.file)
		if err != nil || dir != app || pkg != tc.wantPkg {
			t.Errorf("buildTarget(%s) = %q, %q, %v; want %q, %q", tc.file, dir, pkg, err, app, tc.wantPkg)
		}
	}
	// A run rooted in a subdirectory of its module, mutating a file elsewhere in
	// the same module, keeps the subdirectory as the build directory.
	sub := filepath.Join(app, "sub")
	dir, pkg, err := buildTarget(sub, filepath.Join(app, "a.go"))
	if err != nil || dir != sub || pkg != "./.." {
		t.Errorf("same-module target outside the root = %q, %q, %v; want %q, %q", dir, pkg, err, sub, "./..")
	}
}

func TestBuildTarget_CrossModuleUsesTargetRoot(t *testing.T) {
	app, lib := twoModules(t, "package lib\n")
	dir, pkg, err := buildTarget(app, filepath.Join(lib, "inner", "doc.go"))
	if err != nil || dir != lib || pkg != "./inner" {
		t.Errorf("buildTarget = %q, %q, %v; want %q, %q", dir, pkg, err, lib, "./inner")
	}
	if strings.Contains(pkg, "..") {
		t.Errorf("pattern %q escapes its build directory", pkg)
	}
}
