package cueconfig

import (
	_ "embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
)

func getOverlay(fsys fs.FS) (map[string]load.Source, err) {

	overlay := map[string]load.Source{}
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
				//overlayFilename := fmt.Sprintf(overlayFmt, filesystemID, filename)
				//overlayFilename := filename
				path := filepath.Join("XXX", filename)
				overlay[path] = load.FromBytes(data)
				if cg.Debug {
					log.Printf("        * %v -> %v", filename, path)
				}
				if cg.DumpOverlays {
					os.WriteFile(overlayFilename, data, 0644)
				}
			} else {
				if cg.Debug {
					log.Printf("        * %v", filename)
				}
			}
			return nil
		},
	); err != nil {
		return overlay, fmt.Errorf("walkdir: %v", err)
	}
	return overlay, nil
}

func LoadFS(fsys fs.FS, configFilePath string, schema, defaults []byte, runtime any, dest any) error {
	info, err := fs.Stat(fsys, configFilePath)
	if err != nil {
		return err
	}
	overlay, err := getOverlay(fsys)
	if err != nil {
		return err
	}
	var configInst *build.Instance
	if info.IsDir() {
		configInst = load.Instances([]string{"."}, &load.Config{
			Dir:     configFilePath,
			Overlay: overlay,
		})[0]
	} else {
		configInst = load.Instances([]string{configFilePath}, &load.Config{
			Overlay: overlay,
		})[0]
	}
	if err := configInst.Err; err != nil {
		return fmt.Errorf("cannot load configuration from %q: %v", configFilePath, err)
	}
	ctx := cuecontext.New()

	configVal := ctx.BuildInstance(configInst)
	if err := configVal.Validate(cue.All()); err != nil {
		return fmt.Errorf("invalid configuration from %q: %v", configFilePath, errors.Details(err, nil))
	}

	// Compile the config schema.
	schemaVal := ctx.CompileBytes(schema, cue.Filename("$schema.cue"))
	if err := schemaVal.Err(); err != nil {
		return fmt.Errorf("unexpected error in configuration schema %q: %v", schema, errors.Details(err, nil))
	}
	// Compile the defaults.
	defaultsVal := ctx.CompileBytes(defaults, cue.Filename("$defaults.cue"))
	if err := schemaVal.Err(); err != nil {
		return fmt.Errorf("unexpected error in defaults %q: %v", defaults, errors.Details(err, nil))
	}

	// Unify the user-provided config with the configuration schema.
	configVal = configVal.Unify(schemaVal)
	if err := configVal.Validate(cue.All()); err != nil {
		return fmt.Errorf("error in configuration: %v", errors.Details(err, nil))
	}

	// If there's a runtime value provided, unify it with the runtime field.
	if runtime != nil {
		configVal = configVal.Unify(ctx.Encode(runtime))
		if err := configVal.Validate(cue.All()); err != nil {
			return fmt.Errorf("config schema conflict on runtime values: %v", errors.Details(err, nil))
		}
	}

	// The user layer is now complete. Now finalize it and apply any program-level
	// defaults.
	configVal, err = finalize(ctx, configVal)
	if err != nil {
		return fmt.Errorf("internal error: cannot finalize configuration value: %v", errors.Details(err, nil))
	}
	// Unify with the defaults.
	configVal = configVal.Unify(defaultsVal)
	if err := configVal.Validate(cue.All()); err != nil {
		return fmt.Errorf("config schema error applying program-level defaults: %v", errors.Details(err, nil))
	}

	if err := configVal.Decode(dest); err != nil {
		return fmt.Errorf("cannot decode final configuration: %v", errors.Details(err, nil))
	}
	return nil
}
