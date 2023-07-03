// Package `cueconfig` provides an API designed to make it straightforward
// to use the CUE language (see https://cuelang.org) as a configuration format
// for Go programs.
package cueconfig

import (
	_ "embed"
	"fmt"
	"io/fs"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
)

// Load loads the CUE configuration at the file or CUE package directory
// configFilePath. If the file does not exist, an os.IsNotExist error is returned.
//
// The schema for the user configuration is given in schema.
// Usually this will come from a CUE file embedded in the binary by the caller.
// If it's empty then any configuration data will be allowed.
//
// The user configuration is first read and unified with the schema.
// Then if runtime is non-nil, it is unified with the resulting value,
// thus providing the user configuration with any desired runtime values (e.g.
// environment variables).
//
// Then the result is finalised (applying any user-specified defaults), then
// unified with the CUE contained in the defaults argument.
// This allows the program to specify default values independently
// from the user.
//
// The result is unmarshaled into the Go value pointed to by dest
// using cue.Value.Decode (similar to json.Unmarshal).
func Load(configFilePath string, schema, defaults []byte, runtime any, dest any) error {
	return LoadFS(os.DirFS("."), configFilePath, schema, defaults, runtime, dest)
}

// LoadFS loads the CUE configuration from a given fs.FS, see Load
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

// finalize returns v with all references resolved and all default values selected.
// This is a bit of a hack until similar functionality is implemented inside
// the cue package itself.
func finalize(ctx *cue.Context, v cue.Value) (cue.Value, error) {
	n := v.Syntax(cue.Final())
	var f *ast.File
	switch n := n.(type) {
	case *ast.File:
		f = n
	case ast.Expr:
		var err error
		f, err = astutil.ToFile(n)
		if err != nil {
			return cue.Value{}, fmt.Errorf("cannot convert node to expr: %v", err)
		}
	default:
		return cue.Value{}, fmt.Errorf("unexpected type %#v", n)
	}
	return ctx.BuildFile(f), nil
}
