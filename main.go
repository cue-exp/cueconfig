package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
)

type config struct {
	Foo int            `json:"foo"`
	Bar map[string]Baz `json:"bar"`
}

type Baz struct {
	Blah   string `json:"blah"`
	Foobie []int  `json:"foobie"`
}

func main() {
	os.Exit(Main())
}

func Main() int {
	// The path to the configuration directory. Note that it's a directory
	// so we get all the power of CUE, including multiple files, imports,
	// cue.mod etc.
	configPath := ".exampleconfig"

	var cfg config
	if err := loadConfig(configPath, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "cannot load config from %q: %v\n", configPath, errors.Details(err, nil))
		return 1
	}
	data, _ := json.MarshalIndent(cfg, "", "\t")
	fmt.Printf("%s\n", data)
	return 0
}

var configValuePath = cue.MakePath(cue.Str("config"))

//go:embed schema.cue
var schema string

// loadConfig reads the CUE configuration at the given configPath
// directory, unmarshaling it as JSON into the value pointed to
// by into, which should be a pointer to a struct.
func loadConfig(configPath string, into interface{}) error {
	ctx := cuecontext.New()
	schemaVal := ctx.CompileString(schema, cue.Filename("$schema.cue"))
	if err := schemaVal.Err(); err != nil {
		panic(fmt.Errorf("unexpected error in embedded schema: %v", errors.Details(err, nil)))
	}
	configSchemaVal := schemaVal.LookupPath(configValuePath)
	if err := configSchemaVal.Err(); err != nil {
		panic(fmt.Errorf("cannot find %v: %v", configValuePath, err))
	}

	// Load the configuration, which hasn't yet been unified with the runtime config.
	insts := load.Instances([]string{"."}, &load.Config{
		Dir: configPath,
	})
	configVal := ctx.BuildInstance(insts[0])
	final := configVal.Unify(configSchemaVal)

	// Check that it's all OK.
	if err := final.Validate(
		cue.Attributes(true),
		cue.Definitions(true),
		cue.Hidden(true),
	); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return final.Decode(into)
}
