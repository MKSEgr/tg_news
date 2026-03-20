package migrations

import (
	"embed"
	"io/fs"
	"sort"
)

// Files contains embedded SQL migration assets used during startup.
//
//go:embed *.up.sql
var Files embed.FS

// UpFileNames returns embedded up-migration file names in lexical order.
func UpFileNames() ([]string, error) {
	entries, err := fs.ReadDir(Files, ".")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) >= len(".up.sql") && name[len(name)-len(".up.sql"):] == ".up.sql" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}
