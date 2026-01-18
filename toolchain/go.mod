module github.com/varavelio/vdl/toolchain

go 1.25

require github.com/varavelio/vdl/playground v0.0.0

replace github.com/varavelio/vdl/playground v0.0.0 => ../playground

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/alecthomas/participle/v2 v2.1.4
	github.com/alexflint/go-arg v1.5.1
	github.com/goccy/go-yaml v1.18.0
	github.com/orsinium-labs/enum v1.4.0
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.1
	github.com/stretchr/testify v1.10.0
	github.com/uforg/ufogenkit v0.1.0
	golang.org/x/tools v0.32.0
)

require (
	github.com/alexflint/go-scalar v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
