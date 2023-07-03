package cueconfig

import (
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue/load"
)

func getOverlay(fsys fs.FS) (map[string]load.Source, error) {
	overlay := map[string]load.Source{}
	cwd, err := os.Getwd()
	if err != nil {
		return overlay, err
	}
	if err := fs.WalkDir(
		fsys, ".",
		func(filename string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.Type().IsRegular() {
				return nil
			}
			if strings.HasSuffix(filename, ".cue") {
				data, err := fs.ReadFile(fsys, filename)
				if err != nil {
					return err
				}
				path := filepath.Join(cwd, filename)
				overlay[path] = load.FromBytes(data)
			}
			return nil
		},
	); err != nil {
		return overlay, fmt.Errorf("walkdir: %v", err)
	}
	return overlay, nil
}
