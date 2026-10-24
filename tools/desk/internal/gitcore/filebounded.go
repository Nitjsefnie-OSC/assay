package gitcore

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"io"
)

// FileAtBounded reads only a regular blob at an exact caller-selected revision.
// A large blob is reported before decompression; no filesystem path is followed.
func (r *Repo) FileAtBounded(rev, path string, limit int64) (string, bool, error) {
	tree, e := r.treeAt(rev)
	if e != nil {
		return "", false, e
	}
	f, e := tree.File(path)
	if e != nil {
		return "", false, e
	}
	if f.Mode != filemode.Regular && f.Mode != filemode.Executable {
		return "", false, fmt.Errorf("not a regular blob")
	}
	if limit < 0 || f.Size > limit {
		return "", true, nil
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
