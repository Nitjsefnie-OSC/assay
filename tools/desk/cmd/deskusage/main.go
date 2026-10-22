// Command deskusage stores only reported usage counts locally and prints aggregates.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/medici-finance/assay/tools/desk/internal/cli"
	"github.com/medici-finance/assay/tools/desk/internal/deskkit"
	"github.com/spf13/cobra"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
func run(args []string, out, errout io.Writer) int {
	disabled := false
	checked := false
	guard := func() error {
		checked = true
		e := deskkit.Guard()
		disabled = deskkit.ExitCodeOf(e) == deskkit.ExitDisabled
		return e
	}
	rc := cli.Run(func() *cobra.Command {
		root := cli.NewRoot("deskusage", "Local counts only; never reads transcripts, prompts, stdin or tool output")
		sha, built := deskkit.Version()
		root.Version = fmt.Sprintf("deskusage sourceSHA=%s builtAt=%s", sha, built)
		root.PersistentPreRunE = func(*cobra.Command, []string) error {
			return guard()
		}
		root.Args = cobra.NoArgs
		root.RunE = func(*cobra.Command, []string) error { return deskkit.Refused("choose record or summary") }
		root.AddCommand(recordCommand(), summaryCommand())
		return root
	}, args, cli.Options{IO: cli.IO{Out: out, Err: errout}, UsageExit: deskkit.ExitRefused, ExitCode: func(e error) int {
		if e != nil {
			fmt.Fprintln(errout, "deskusage:", e)
		}
		return deskkit.ExitCodeOf(e)
	}})
	// Guard owns the disabled audit entry. Help/version remain side-effect-free.
	help := !checked && rc == 0
	if !checked && !help {
		if e := guard(); e != nil {
			fmt.Fprintln(errout, "deskusage:", e)
			rc = deskkit.ExitCodeOf(e)
		}
	}
	if !disabled && !help {
		verb := "other"
		if len(args) > 0 && (args[0] == "record" || args[0] == "summary") {
			verb = args[0]
		}
		result := deskkit.ResultLocal
		if rc == 5 {
			result = deskkit.ResultRefused
		} else if rc != 0 {
			result = deskkit.ResultUnverifiable
		}
		if e := deskkit.Log(deskkit.Entry{Tool: "deskusage", Verb: verb, Result: result, Detail: "local usage counts"}); e != nil {
			fmt.Fprintln(errout, "deskusage: cannot audit local usage action")
			return deskkit.ExitUnverifiable
		}
	}
	return rc
}
