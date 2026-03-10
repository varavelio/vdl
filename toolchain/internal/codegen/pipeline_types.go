package codegen

import (
	"net/http"

	"github.com/varavelio/vdl/toolchain/internal/codegen/configtypes"
	"github.com/varavelio/vdl/toolchain/internal/codegen/plugintypes"
)

type pluginSourceKind string

const (
	pluginSourceKindLocal  pluginSourceKind = "local"
	pluginSourceKindRemote pluginSourceKind = "remote"
)

type pluginSource struct {
	Kind         pluginSourceKind
	Original     string
	DisplayName  string
	LocalPath    string
	CanonicalURL string
	AuthMatchURL string
	Headers      http.Header
	CachePath    string
	ContentHash  string
}

type runtimePlugin struct {
	Index          int
	Source         pluginSource
	SchemaPath     string
	OutDir         string
	Options        map[string]string
	GenerateHeader bool
}

type preparedPlugin struct {
	Plugin runtimePlugin
	Script string
	Input  plugintypes.PluginInput
}

type executedPlugin struct {
	Plugin runtimePlugin
	Output plugintypes.PluginOutput
}

type outputWrite struct {
	PluginName   string
	OutDir       string
	RelativePath string
	AbsolutePath string
	Content      string
}

type outputPlan struct {
	OutDirs []string
	Writes  []outputWrite
}

type remoteHostAuth struct {
	Host string
	Auth configtypes.VdlConfigRemoteAuth
}
