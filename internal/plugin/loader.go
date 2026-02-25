package plugin

import (
	"fmt"
	"io/fs"
)

// Loader handles discovering and loading species packs from a filesystem.
type Loader struct{}

// NewLoader creates a new Loader.
func NewLoader() *Loader {
	return &Loader{}
}

// LoadAll discovers and loads all species packs from the given filesystem root.
// Each immediate subdirectory of root that contains a species.toml is treated
// as a species pack.
func (l *Loader) LoadAll(fsys fs.FS, root string, source PluginSource) ([]*SpeciesPack, error) {
	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil, fmt.Errorf("read plugin directory %q: %w", root, err)
	}

	var packs []*SpeciesPack
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := root + "/" + entry.Name()
		if root == "." {
			dir = entry.Name()
		}

		// Check if species.toml exists
		if _, err := fs.Stat(fsys, dir+"/species.toml"); err != nil {
			continue // not a species pack directory, skip
		}

		pack, err := ParsePack(fsys, dir)
		if err != nil {
			return nil, fmt.Errorf("load pack %q: %w", entry.Name(), err)
		}
		pack.Source = source

		// Validate
		if errs := Validate(pack); len(errs) > 0 {
			var msg string
			for _, e := range errs {
				msg += "\n  - " + e.Error()
			}
			return nil, fmt.Errorf("validate pack %q:%s", entry.Name(), msg)
		}

		packs = append(packs, pack)
	}

	return packs, nil
}

// LoadOne loads a single species pack from the given directory.
func (l *Loader) LoadOne(fsys fs.FS, dir string, source PluginSource) (*SpeciesPack, error) {
	pack, err := ParsePack(fsys, dir)
	if err != nil {
		return nil, err
	}
	pack.Source = source

	if errs := Validate(pack); len(errs) > 0 {
		var msg string
		for _, e := range errs {
			msg += "\n  - " + e.Error()
		}
		return nil, fmt.Errorf("validate pack:%s", msg)
	}

	return pack, nil
}
