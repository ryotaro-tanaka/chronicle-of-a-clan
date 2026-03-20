package datafiles

import (
	"path/filepath"
	"runtime"
)

func Path(rel string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return rel
	}
	root := filepath.Join(filepath.Dir(filename), "..", "..", "..")
	return filepath.Join(root, rel)
}
