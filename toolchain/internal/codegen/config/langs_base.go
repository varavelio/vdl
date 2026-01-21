package config

// BaseCodeOptions defines standard options shared across code generators.
type BaseCodeOptions struct {
	GenClient   bool  `yaml:"gen_client" json:"gen_client,omitempty" jsonschema:"default=false,description=Generate RPC client code."`
	GenServer   bool  `yaml:"gen_server" json:"gen_server,omitempty" jsonschema:"default=false,description=Generate RPC server interfaces and handlers."`
	GenPatterns *bool `yaml:"gen_patterns" json:"gen_patterns,omitempty" jsonschema:"default=true,description=Generate helper functions for patterns."`
	GenConsts   *bool `yaml:"gen_consts" json:"gen_consts,omitempty" jsonschema:"default=true,description=Generate constant definitions."`
}

// ShouldGenPatterns returns true if patterns should be generated (default: true).
func (b BaseCodeOptions) ShouldGenPatterns() bool {
	if b.GenPatterns == nil {
		return true
	}
	return *b.GenPatterns
}

// ShouldGenConsts returns true if constants should be generated (default: true).
func (b BaseCodeOptions) ShouldGenConsts() bool {
	if b.GenConsts == nil {
		return true
	}
	return *b.GenConsts
}
