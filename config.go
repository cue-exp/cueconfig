package cueconfig

import (
	_ "embed"
	"fmt"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
)

//go:embed metaschema.cue
var metaSchema string

var (
	pathConfig       = cue.MakePath(cue.Def("#Config"))
	pathRuntime      = cue.MakePath(cue.Def("#Runtime"))
	pathDefaults     = cue.MakePath(cue.Def("#Defaults"))
	pathRuntimeField = cue.MakePath(cue.Str("runtime"))
	pathConfigField  = cue.MakePath(cue.Str("config"))
)

// Load loads the CUE configuration at the file or CUE package directory
// configFilePath. If the file does not exist, os.IsNotExist error is returned.
//
// The schema for the user configuration is given in schemaBytes.
// Usually this will come from a CUE file embedded in the binary by the caller.
// If it's empty then any configuration data will be allowed.
//
// The result is unmarshaled into the Go value pointed to by dest
// using cue.Value.Decode (similar to json.Unmarshal).
//
// The schema must conform to the following meta-schema:
//
//	// #Config defines the configuration schema that the user must specify
//	// in their configuration file. It should be a pure schema with no defaults.
//	// Note that the default is to allow any configuration at all.
//	#Config: {
//		...
//	}
//
//	// #Runtime holds runtime values that will be mixed into the configuration
//	// in addition to the user-specified configuration. Examples might
//	// be environment variables or the current working directory.
//	#Runtime: _
//
//	// #Defaults holds any program-defined default values
//	// for the configuration. Any defaults supplied by the user's
//	// configuration will have been resolved before this is
//	// applied.
//	//
//	// Note that if this is not supplied, there will be no program-defined
//	// defaults filled in by Load.
//	#Defaults: {
//		// runtime holds any values supplied as part of the runtime
//		// parameter to Load.
//		runtime: #Runtime
//		// config should define any default values, possibly in terms
//		// of the runtime values.
//		config: #Config
//	}
func Load(configFilePath string, schemaBytes []byte, runtime any, dest any) error {
	info, err := os.Stat(configFilePath)
	if err != nil {
		return err
	}
	var configInst *build.Instance
	if info.IsDir() {
		configInst = load.Instances([]string{"."}, &load.Config{
			Dir: configFilePath,
		})[0]
	} else {
		configInst = load.Instances([]string{configFilePath}, nil)[0]
	}
	if err := configInst.Err; err != nil {
		return fmt.Errorf("cannot load configuration from %q: %v", configFilePath, err)
	}
	ctx := cuecontext.New()

	configVal := ctx.BuildInstance(configInst)
	if err := configVal.Validate(cue.All()); err != nil {
		return fmt.Errorf("invalid configuration from %q: %v", configFilePath, errors.Details(err, nil))
	}

	metaSchemaVal := ctx.CompileString(metaSchema, cue.Filename("$metaschema.cue"))
	if err := metaSchemaVal.Err(); err != nil {
		panic(fmt.Errorf("unexpected error in meta schema: %v", errors.Details(err, nil)))
	}

	schemaVal := ctx.CompileBytes(schemaBytes, cue.Filename("$schema.cue"))
	if err := schemaVal.Err(); err != nil {
		return fmt.Errorf("unexpected error in config schema %q: %v", schemaBytes, errors.Details(err, nil))
	}

	// Unify the metaschema with the actual configuration schema.
	schemaVal = schemaVal.Unify(metaSchemaVal)
	if err := schemaVal.Validate(cue.All()); err != nil {
		return fmt.Errorf("config schema conflict: %v", errors.Details(err, nil))
	}

	// Unify the schema with the user-provided config.
	configVal = configVal.Unify(schemaVal.LookupPath(pathConfig))
	if err := configVal.Validate(cue.All()); err != nil {
		return fmt.Errorf("error(s) in configuration: %v", errors.Details(err, nil))
	}

	// finalize the configuration as supplied by the user, so that any defaults
	// they use won't conflict with defaults supplied by the program.
	configVal, err = finalize(ctx, configVal)
	if err != nil {
		return fmt.Errorf("internal error: cannot finalize configuratiuon value: %v", errors.Details(err, nil))
	}

	// Get the #Defaults definition.
	defaults := schemaVal.LookupPath(pathDefaults)

	// Fill in the runtime field with the actual runtime values.
	defaults = defaults.FillPath(pathRuntimeField, runtime)
	if err := defaults.Validate(cue.All()); err != nil {
		return fmt.Errorf("error in program-supplied runtime values: %v", err)
	}

	// Unify the #Defaults.config field with the finalized configuration.
	defaults = defaults.FillPath(pathConfigField, configVal)
	if err := defaults.Validate(cue.All()); err != nil {
		return fmt.Errorf("cannot fill in defaults: %v", errors.Details(err, nil))
	}

	// Read out the final configuration value with all defaults applied.
	configVal = defaults.LookupPath(pathConfigField)
	if err := configVal.Decode(dest); err != nil {
		return fmt.Errorf("cannot decode final configuration: %v", errors.Details(err, nil))
	}
	return nil
}

// finalize returns v with all references resolved and all default values selected.
// This is a bit of a hack until similar functionality is implemented inside
// the cue package itself.
func finalize(ctx *cue.Context, v cue.Value) (cue.Value, error) {
	n := v.Syntax(cue.Final(), cue.ResolveReferences(true), cue.DisallowCycles(true))
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
