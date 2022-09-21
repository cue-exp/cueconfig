package cueconfig_test

import (
	"encoding/json"
	"fmt"

	"github.com/cue-exp/cueconfig"
)

type exampleConfig struct {
	Foo int                    `json:"foo"`
	Bar map[string]*exampleBar `json:"bar"`
}

type exampleBar struct {
	Amount    float64 `json:"amount"`
	Something bool    `json:"something"`
	Path      string  `json:"path"`
}

type exampleRuntime struct {
	CurrentDirectory string `json:"currentDirectory"`
}

func ExampleLoad() {
	// In this example, we use the example.cue from the package
	// but in practice this would be located whereever you'd want
	// your program's configuration file.
	configFile := "example.cue"

	// This is a placeholder for any runtime values provided
	// as input to the configuration.
	runtime := struct {
		Runtime exampleRuntime `json:"runtime"`
	}{
		Runtime: exampleRuntime{
			CurrentDirectory: "/path/to/current/directory",
		},
	}
	// Load the configuration into the Go value cfg.
	var cfg exampleConfig
	if err := cueconfig.Load(configFile, exampleSchema, exampleDefaults, runtime, &cfg); err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	// This is a placeholder for anything that the program might actually do
	// with the configuration.
	data, _ := json.MarshalIndent(cfg, "", "\t")
	fmt.Printf("%s\n", data)
	//Output:
	//{
	//	"foo": 1,
	//	"bar": {
	//		"a": {
	//			"amount": 1.5,
	//			"something": false,
	//			"path": "/path/to/current/directory"
	//		}
	//	}
	//}
}

// exampleSchema holds the schema for the program's configuration.
// It would be conventional to include it as an embedded file,
// but we've included it inline here so that it shows up directly
// in the example.
var exampleSchema = []byte(`
package example

foo?: int
bar?: [string]: #Bar

#Bar: {
	amount?: number
	something?: bool
	path?: string
}

// runtime is provided by the program, not the user.
runtime?: #Runtime
#Runtime: {
	currentDirectory: string
}
`)

// exampleDefaults holds any default values for the program
// to apply. As with [exampleSchema] above, it would conventionally
// be taken from an embedded file.
var exampleDefaults = []byte(`
runtime: _
foo: *100 | _
bar: [_]: {
	amount: *1.5 | _
	something: *false | _
	path: *runtime.currentDirectory | _
}
`)
