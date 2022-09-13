package cueconfig_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/rogpeppe/cueconfig"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestFoo(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:           "testdata",
		UpdateScripts: os.Getenv("SCRIPT_UPDATE") != "",
	})
}

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"cueconfig-test": Main,
	}))
}

//go:embed testschema.cue
var schema []byte

type config struct {
	Foo int            `json:"foo"`
	Bar map[string]Baz `json:"bar"`
}

type Baz struct {
	Blah   string `json:"blah"`
	Foobie []int  `json:"foobie"`
}

func Main() int {
	var cfg config
	if err := cueconfig.Load(".exampleconfig", schema, nil, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	data, _ := json.MarshalIndent(cfg, "", "\t")
	fmt.Printf("%s\n", data)
	return 0
}
