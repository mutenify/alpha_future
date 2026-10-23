// Package storage — файлы документов на локальном диске (том в docker compose).
package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type Local struct{ root string }

func NewLocal(root string) *Local { return &Local{root: root} }

// Save: Clean от корня не даёт ключу выйти за root.
func (l *Local) Save(_ context.Context, key string, r io.Reader) (int64, error) {
	path := filepath.Join(l.root, filepath.Clean("/"+key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, err
	}
	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}
