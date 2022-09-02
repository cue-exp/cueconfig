// This file defines the "meta-schema" - a schema that outlines the schema that can be provided
// by packages that use the cueconfig package.

// #Config defines the configuration schema that the user must specify
// in their configuration file. It should be a pure schema with no defaults.
// Note that the default is to allow any configuration at all.
#Config: {
	...
}

// #Runtime holds runtime values that will be mixed into the configuration
// in addition to the user-specified configuration. Examples might
// be environment variables or the current working directory.
#Runtime: {
	...
}

// #Defaults holds any program-defined default values
// for the configuration. Any defaults supplied by the user's
// configuration will have been resolved before this is
// applied.
//
// Note that if this is not supplied, there will be no program-defined
// defaults filled in by Load.
#Defaults: {
	// runtime holds any values supplied as part of the runtime
	// parameter to Load.
	runtime: #Runtime
	// config should define any default values, possibly in terms
	// of the runtime values.
	config: #Config
}
