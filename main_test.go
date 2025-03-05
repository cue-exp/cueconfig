package cueconfig_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cue-exp/cueconfig"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestFoo(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:           "testdata",
		UpdateScripts: os.Getenv("SCRIPT_UPDATE") != "",
	})
}

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"cueconfig-test": Main,
	})
}

//go:embed testschema.cue
var schema []byte

//go:embed testdefaults.cue
var defaults []byte

type config struct {
	Foo int            `json:"foo"`
	Bar map[string]Baz `json:"bar"`
}

type Baz struct {
	Blah   string `json:"blah"`
	Foobie []int  `json:"foobie"`
}

func Main() {
	configFile := ".exampleconfig"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	runtime := struct {
		Env map[string]string `json:"env"`
	}{environ()}
	var cfg config
	if err := cueconfig.Load(configFile, schema, defaults, runtime, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	data, _ := json.MarshalIndent(cfg, "", "\t")
	fmt.Printf("%s\n", data)
}

func environ() map[string]string {
	m := make(map[string]string)
	for _, e := range os.Environ() {
		k, v, _ := strings.Cut(e, "=")
		m[k] = v
	}
	return m
}
