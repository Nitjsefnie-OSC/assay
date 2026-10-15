package main

// cli.go is statusgen's command-line adapter: the Cobra/pflag/Viper wiring every statusgen
// command runs through (desktools-v2 cli-contract.md). It is a statusgen-local copy of the
// pattern the desk module's internal/cli package implements, because statusgen is its own
// module and release asset and must not import desk internals.
//
// The adapter parses, validates and answers help and version; it is never an authority.
// Every handler keeps its own domain checks, exit codes and environment reads.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// cliExitUsage is the exit code for a parse error: an unknown flag or command, a missing or
// malformed flag value, or a positional the command refuses. It is the code every statusgen
// verb's Go-flag parser used before the Cobra tree.
const cliExitUsage = 2

// cliIO is the three streams one invocation reads and writes.
type cliIO struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// processIO is the process's own streams, for main and the verb wrappers tests call.
func processIO() cliIO { return cliIO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr} }

// exitCodeError carries a handler's non-zero exit code out of a RunE.
type exitCodeError struct{ code int }

func (e exitCodeError) Error() string { return fmt.Sprintf("exit status %d", e.code) }
func (e exitCodeError) ExitCode() int { return e.code }

// codeErr turns a handler's exit code into its RunE result.
func codeErr(code int) error {
	if code == 0 {
		return nil
	}
	return exitCodeError{code}
}

// verbHandler adapts a handler that returns an exit code to a Cobra RunE.
func verbHandler(h func(args []string) int) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error { return codeErr(h(args)) }
}

// newCommand returns a command with the adapter's fixed settings: Cobra's own error and
// usage printing off (runCLI prints a parse error once; a handler reports its own errors).
func newCommand(use, short, long string) *cobra.Command {
	return &cobra.Command{
		Use:               use,
		Short:             short,
		Long:              long,
		SilenceErrors:     true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
}

// handlerError marks an error returned by a handler, so runCLI can tell it from a parse error.
type handlerError struct{ err error }

func (h handlerError) Error() string { return h.err.Error() }
func (h handlerError) Unwrap() error { return h.err }

// preParseCheck is a refusal a command makes over its raw arguments before they are parsed,
// so its message is not replaced by the parser's generic unknown-flag error. It returns 0
// to let the line through. Keyed by command path; the map is never written.
var preParseChecks = map[string]func(args []string, stderr io.Writer) int{
	"statusgen verifyrun": verifyrunRefuseRunnerFlags,
}

// runCLI builds a fresh tree with build, executes args against it and returns the exit code.
// It reads no configuration, roster or credential itself: help and version are answered by
// the tree before any handler runs.
func runCLI(build func(cliIO) *cobra.Command, args []string, rio cliIO) int {
	if rio.In == nil {
		rio.In = strings.NewReader("")
	}
	if rio.Out == nil {
		rio.Out = io.Discard
	}
	if rio.Err == nil {
		rio.Err = io.Discard
	}
	root := build(rio)
	root.SetIn(rio.In)
	root.SetOut(rio.Out)
	root.SetErr(rio.Err)
	root.SilenceErrors = true
	root.SilenceUsage = true
	root.CompletionOptions.DisableDefaultCmd = true
	// Cobra's help subcommand is runnable, unlike --help. Identify it before wrapping so the
	// settings resolution never runs for it.
	root.InitDefaultHelpCmd()
	help, _, helpErr := root.Find([]string{"help"})
	if helpErr != nil || help == root {
		help = nil
	}
	wrapHandlers(root, help)
	refuseCompletionRequests(root)
	if args == nil {
		args = []string{}
	}
	if found, rest, err := root.Find(args); err == nil {
		if check, ok := preParseChecks[found.CommandPath()]; ok {
			if code := check(rest, rio.Err); code != 0 {
				return code
			}
		}
	}
	args = normalizeGoFlags(root, args)
	root.SetArgs(args)
	cmd, err := root.ExecuteC()
	if err == nil {
		return 0
	}
	var h handlerError
	if errors.As(err, &h) {
		var coded interface{ ExitCode() int }
		if errors.As(h.err, &coded) {
			return coded.ExitCode()
		}
		fmt.Fprintf(rio.Err, "%s: %v\n", cmd.CommandPath(), h.err)
		return 1
	}
	path := root.Name()
	if cmd != nil {
		path = cmd.CommandPath()
	}
	fmt.Fprintf(rio.Err, "%s: %v\nRun '%s --help' for usage.\n", path, err, path)
	return cliExitUsage
}

// runVerb runs one verb's command line through the same tree and parser the binary uses:
// args are the verb's own arguments (the verb word excluded). The verb wrappers the tests
// call (runMergecheck and friends) are this, so a test exercises the production parse.
func runVerb(build func(cliIO) *cobra.Command, args []string, rio cliIO) int {
	name := strings.Fields(build(cliIO{}).Use)[0]
	tree := func(rio cliIO) *cobra.Command {
		root := newCommand("statusgen", "", "")
		root.Args = cobra.NoArgs
		root.AddCommand(build(rio))
		return root
	}
	return runCLI(tree, append([]string{name}, args...), rio)
}

// fileOr is w when it is an *os.File, else def: the handlers that hand their streams to
// file-typed helpers keep the process streams when a test passes a buffer.
func fileOr(w io.Writer, def *os.File) *os.File {
	if f, ok := w.(*os.File); ok {
		return f
	}
	return def
}

// settingsKey is the context key the resolved settings travel under to the handler.
type settingsKey struct{}

// setting is one resolved declared setting: its value as the handler reads it and the
// source it came from ("flag" or "default"; statusgen binds no other source).
type setting struct {
	Value  any
	Source string
}

// resolvedSettings returns the settings resolveSettings attached to cmd's context.
func resolvedSettings(cmd *cobra.Command) map[string]setting {
	if ctx := cmd.Context(); ctx != nil {
		if s, ok := ctx.Value(settingsKey{}).(map[string]setting); ok {
			return s
		}
	}
	return nil
}

// resolveSettings resolves every declared flag of cmd through a fresh, command-local Viper
// instance: the flag layer when the flag appears on the line (an explicit empty value
// included), else the flag's default. Nothing else is bound: no environment variable, no
// config file, no global instance. A value the resolution changes is written back to the
// flag, so the handler reads what Viper resolved.
func resolveSettings(cmd *cobra.Command) error {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))
	type decl struct {
		f     *pflag.Flag
		slice bool
	}
	decls := map[string]decl{}
	var keys []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" || f.Name == "version" {
			return
		}
		_, slice := f.Value.(pflag.SliceValue)
		decls[f.Name] = decl{f, slice}
		keys = append(keys, f.Name)
		if slice {
			v.SetDefault(f.Name, []string{})
			if f.Changed {
				v.Set(f.Name, append([]string(nil), f.Value.(pflag.SliceValue).GetSlice()...))
			}
			return
		}
		v.SetDefault(f.Name, f.DefValue)
		if f.Changed {
			v.Set(f.Name, f.Value.String())
		}
	})
	sort.Strings(keys)
	out := make(map[string]setting, len(keys))
	for _, k := range keys {
		d := decls[k]
		src := "default"
		if d.f.Changed {
			src = "flag"
		}
		if d.slice {
			got := v.GetStringSlice(k)
			if cur := d.f.Value.(pflag.SliceValue).GetSlice(); !reflect.DeepEqual(nonNil(got), nonNil(cur)) {
				if err := d.f.Value.(pflag.SliceValue).Replace(got); err != nil {
					return fmt.Errorf("--%s: %v", k, err)
				}
			}
			out[k] = setting{Value: nonNil(got), Source: src}
			continue
		}
		got := v.GetString(k)
		if got != d.f.Value.String() {
			if err := d.f.Value.Set(got); err != nil {
				return fmt.Errorf("--%s: invalid value from %s", k, src)
			}
		}
		out[k] = setting{Value: got, Source: src}
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.SetContext(context.WithValue(ctx, settingsKey{}, out))
	return nil
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// settingsError marks a resolution failure: a usage error, reported like a parse error.
type settingsError struct{ err error }

func (s settingsError) Error() string { return s.err.Error() }

// wrapHandlers resolves each command's settings before its handler and marks every handler
// error in the tree. Argument validation and flag parsing are not handlers.
func wrapHandlers(c, help *cobra.Command) {
	if c.RunE != nil && c != help {
		run := c.RunE
		opaque := c.DisableFlagParsing
		c.RunE = func(cmd *cobra.Command, args []string) error {
			if !opaque {
				if err := resolveSettings(cmd); err != nil {
					return settingsError{err}
				}
			}
			if err := run(cmd, args); err != nil {
				return handlerError{err}
			}
			return nil
		}
	}
	for _, sub := range c.Commands() {
		wrapHandlers(sub, help)
	}
}

// refuseCompletionRequests makes Cobra's hidden shell-completion entrypoints (__complete and
// __completeNoDesc) refuse like any other unknown command: statusgen offers no completion.
func refuseCompletionRequests(root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:                cobra.ShellCompRequestCmd,
		Aliases:            []string{cobra.ShellCompNoDescRequestCmd},
		Hidden:             true,
		DisableFlagParsing: true,
		Args: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("unknown command %q for %q", cmd.CalledAs(), root.CommandPath())
		},
		Run: func(*cobra.Command, []string) {},
	})
}

// normalizeGoFlags rewrites the Go flag package's single-dash long spellings (-name,
// -name=value) to --name forms for the long flags of the command each word is parsed
// against (plus help and version), so every caller written against the old parser keeps
// working. It leaves values, shorthands and everything after "--" untouched, copies the
// arguments of a command that parses no flags verbatim, and stops at the root's first
// positional: the root parses as Go's flag package did, stopping at the first non-flag, so
// what follows it (--export-evidence's `-o`/`-generated` scan) is the handler's.
func normalizeGoFlags(root *cobra.Command, args []string) []string {
	cur := root
	positional := false
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			return append(out, args[i:]...)
		}
		if !strings.HasPrefix(a, "-") || a == "-" {
			if sub := subcommandNamed(cur, a); sub != nil && !positional {
				cur = sub
				out = append(out, a)
				if sub.DisableFlagParsing {
					return append(out, args[i+1:]...)
				}
				continue
			}
			if cur == root {
				return append(out, args[i:]...)
			}
			positional = true
			out = append(out, a)
			continue
		}
		rewritten := a
		if !strings.HasPrefix(a, "--") {
			n := a[1:]
			if j := strings.IndexByte(n, '='); j >= 0 {
				n = n[:j]
			}
			if f := longFlag(cur, n); f != nil && (len(n) > 1 || cur.Flags().ShorthandLookup(n) == nil) {
				rewritten = "-" + a
			}
		}
		out = append(out, rewritten)
		if i+1 >= len(args) || strings.Contains(rewritten, "=") {
			continue
		}
		var f *pflag.Flag
		if strings.HasPrefix(rewritten, "--") {
			f = longFlag(cur, rewritten[2:])
		} else if len(rewritten) == 2 {
			f = cur.Flags().ShorthandLookup(rewritten[1:])
		}
		if f != nil && f.NoOptDefVal == "" {
			out = append(out, args[i+1])
			i++
		}
	}
	return out
}

// longFlag is the long flag named n that c parses (help and version included), or nil.
func longFlag(c *cobra.Command, n string) *pflag.Flag {
	if n == "help" {
		return &pflag.Flag{Name: "help", NoOptDefVal: "true"}
	}
	if f := c.Flags().Lookup(n); f != nil {
		return f
	}
	return c.InheritedFlags().Lookup(n)
}

// subcommandNamed is the child of c named (or aliased) a, nil when there is none.
func subcommandNamed(c *cobra.Command, a string) *cobra.Command {
	for _, s := range c.Commands() {
		if s.Name() == a || s.HasAlias(a) {
			return s
		}
	}
	return nil
}
