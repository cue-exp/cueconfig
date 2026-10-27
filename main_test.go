package cueconfig_test

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

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

func TestMissingConfig(t *testing.T) {
	// A configuration file which does not exist gives a not-exist error,
	// but dest is still filled with the defaults.
	missing := filepath.Join(t.TempDir(), "missing.cue")
	tests := []struct {
		name string
		load func(defaults []byte, dest any) error
	}{
		{"Load", func(defaults []byte, dest any) error {
			return cueconfig.Load(missing, schema, defaults, nil, dest)
		}},
		{"LoadFS", func(defaults []byte, dest any) error {
			return cueconfig.LoadFS(fstest.MapFS{}, "missing.cue", schema, defaults, nil, dest)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var cfg config
			err := test.load(defaults, &cfg)
			if !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("got error %v, want a not-exist error", err)
			}
			want := config{Foo: 1, Bar: map[string]Baz{}}
			if !reflect.DeepEqual(cfg, want) {
				t.Errorf("got %#v, want %#v", cfg, want)
			}

			// If applying the defaults fails, that error is returned instead,
			// still mentioning the missing file.
			err = test.load([]byte("foo: 200"), &cfg)
			if err == nil || errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "missing.cue") {
				t.Fatalf("got error %v, want an error from the defaults mentioning the missing file", err)
			}
		})
	}
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
