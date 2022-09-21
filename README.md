# cueconfig: use CUE to configure your Go programs

Package `cueconfig` provides an API designed to make it straightforward
to use the [CUE language](https://cuelang.org) as a configuration format
for Go programs.

It provides a single entry point, [Load](https://pkg.go.dev/github.com/cue-exp/cueconfig#Load), which is very roughly equivalent to [json.Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal). It loads a configuration file (or directory) from disk and unmarshals it into a Go value.

Optionally, `Load` can be provided with a CUE schema to verify the configuration before unmarshaling into the Go value, a set of default values to apply after any user-specified defaults, and some runtime-defined values to be made available to the configuration.

There is a [full example](https://pkg.go.dev/github.com/cue-exp/cueconfig#example-Load) in the package documentation.
