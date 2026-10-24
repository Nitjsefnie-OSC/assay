package gitcore

import (
	"context"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// FileAtBounded reads only a regular blob at an exact caller-selected revision.
// A large blob is reported before decompression; no filesystem path is followed.
func (r *Repo) FileAtBounded(rev, path string, limit int64) (string, bool, error) {
	tree, e := r.treeAt(rev)
	if e != nil {
		return "", false, e
	}
	entry, e := tree.FindEntry(path)
	if e != nil {
		return "", false, e
	}
	if entry.Mode != filemode.Regular && entry.Mode != filemode.Executable {
		return "", false, fmt.Errorf("not a regular blob")
	}
	size, e := tree.Size(path)
	if e != nil {
		return "", false, e
	}
	if limit < 0 || size > limit {
		return "", true, nil
	}
	f, e := tree.File(path)
	if e != nil {
		return "", false, e
	}
	rd, e := f.Reader()
	if e != nil {
		return "", false, e
	}
	defer rd.Close()
	b, e := io.ReadAll(io.LimitReader(rd, limit+1))
	if int64(len(b)) > limit {
		return "", true, nil
	}
	return string(b), false, e
}

// DiffSymmetricBounded checks changed blob headers before rename detection or
// patch construction can inflate their contents. Over-budget inputs are omitted;
// no unbounded provider fallback is implied by this result.
func (r *Repo) DiffSymmetricBounded(a, b string, contextLines int, budget int64) (string, bool, error) {
	base, e := r.MergeBase(a, b)
	if e != nil {
		return "", false, e
	}
	from, e := r.treeAt(base)
	if e != nil {
		return "", false, e
	}
	to, e := r.treeAt(b)
	if e != nil {
		return "", false, e
	}
	changes, e := object.DiffTreeWithOptions(context.Background(), from, to, &object.DiffTreeOptions{DetectRenames: false})
	if e != nil {
		return "", false, e
	}
	seen := map[plumbing.Hash]bool{}
	var total int64
	for _, c := range changes {
		for _, v := range []object.ChangeEntry{c.From, c.To} {
			if v.Tree == nil || v.TreeEntry.Hash.IsZero() || seen[v.TreeEntry.Hash] {
				continue
			}
			if v.TreeEntry.Mode != filemode.Regular && v.TreeEntry.Mode != filemode.Executable {
				return "", true, nil
			}
			seen[v.TreeEntry.Hash] = true
			n, e := r.repo.Storer.EncodedObjectSize(v.TreeEntry.Hash)
			if e != nil {
				return "", false, e
			}
			if n < 0 || n > budget-total {
				return "", true, nil
			}
			total += n
		}
	}
	raw, e := r.Diff(base, b, contextLines)
	return raw, false, e
}
