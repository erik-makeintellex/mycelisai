package configdocuments

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSeedBuiltInRevisionsCallersAreBootstrapOnly keeps the only built-in
// writer reachable from Core startup alone.
func TestSeedBuiltInRevisionsCallersAreBootstrapOnly(t *testing.T) {
	coreRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		filepath.Join("internal", "configdocuments"): true,
		filepath.Join("cmd", "server"):               true,
	}
	var offenders []string
	err = filepath.WalkDir(coreRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if name := entry.Name(); name == "vendor" || name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(raw), "SeedBuiltInRevisions") {
			return nil
		}
		rel, err := filepath.Rel(coreRoot, filepath.Dir(path))
		if err != nil {
			return err
		}
		if !allowed[rel] {
			offenders = append(offenders, filepath.Join(rel, filepath.Base(path)))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(offenders) != 0 {
		t.Fatalf("SeedBuiltInRevisions referenced outside bootstrap: %v", offenders)
	}
}
