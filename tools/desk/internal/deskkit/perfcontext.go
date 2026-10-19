package deskkit

import (
	"context"

	"github.com/medici-finance/assay/perfrec"
)

type perfContextKey struct{}

// WithPerfContext carries immutable diagnostics beside an explicit call, never
// in package-global work state or operational authority.
func WithPerfContext(ctx context.Context, c perfrec.Context) context.Context {
	return context.WithValue(ctx, perfContextKey{}, c)
}
func PerfContext(ctx context.Context) (perfrec.Context, bool) {
	c, ok := ctx.Value(perfContextKey{}).(perfrec.Context)
	return c, ok
}
func PerfContextEnv(env []string, c perfrec.Context, toolChild bool) []string {
	return perfrec.ContextEnv(env, c, toolChild)
}
func PerfAgentContext(parent perfrec.Context, role string, work []perfrec.WorkRef, resume *perfrec.Attribution) perfrec.Context {
	return perfrec.ChildContext(parent, role, work, resume)
}
