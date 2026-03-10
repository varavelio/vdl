package codegen

import (
	"encoding/base64"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/varavelio/vdl/toolchain/internal/codegen/configtypes"
)

const expectedConfigVersion int64 = 1

var githubPluginPattern = regexp.MustCompile(`^([^/\s]+)/([^@/\s]+)@([^\s]+)$`)

// resolveRuntimePlugins validates the high-level generator configuration and
// normalizes every configured plugin into a runtime-ready description.
func resolveRuntimePlugins(config runtimeConfig) ([]runtimePlugin, error) {
	if config.Config.GetVersion() != expectedConfigVersion {
		return nil, fmt.Errorf(
			"unsupported config version %d in %q: expected %d",
			config.Config.GetVersion(),
			config.Path,
			expectedConfigVersion,
		)
	}

	remoteAuths, err := normalizeRemoteAuths(config.Config.GetRemotes())
	if err != nil {
		return nil, err
	}

	pluginsConfig := config.Config.GetPlugins()
	plugins := make([]runtimePlugin, 0, len(pluginsConfig))
	for i, pluginConfig := range pluginsConfig {
		source, err := resolvePluginSource(config.Dir, pluginConfig.GetSrc(), remoteAuths)
		if err != nil {
			return nil, fmt.Errorf("plugin %d: %w", i+1, err)
		}

		schemaPath, err := resolveAbsolutePath(config.Dir, pluginConfig.GetSchema(), ".vdl", "schema")
		if err != nil {
			return nil, fmt.Errorf("plugin %d: %w", i+1, err)
		}

		outDir, err := resolveAbsolutePath(config.Dir, pluginConfig.GetOutDir(), "", "outDir")
		if err != nil {
			return nil, fmt.Errorf("plugin %d: %w", i+1, err)
		}

		plugins = append(plugins, runtimePlugin{
			Index:          i,
			Source:         source,
			SchemaPath:     schemaPath,
			OutDir:         outDir,
			Options:        cloneStringMap(pluginConfig.GetOptionsOr(map[string]string{})),
			GenerateHeader: configtypes.Or(pluginConfig.GenerateHeader, true),
		})
	}

	return plugins, nil
}

// normalizeRemoteAuths validates configured remotes and converts them into the
// host matcher representation used during remote plugin resolution.
func normalizeRemoteAuths(remotes []configtypes.VdlConfigRemote) ([]remoteHostAuth, error) {
	result := make([]remoteHostAuth, 0, len(remotes))
	for i, remote := range remotes {
		host := normalizeRemoteHost(remote.GetHost())
		if host == "" {
			return nil, fmt.Errorf("remote %d has an empty host", i+1)
		}

		if strings.Contains(host, "://") {
			return nil, fmt.Errorf("remote %q must not include a URL scheme", remote.GetHost())
		}

		if remote.Auth == nil {
			continue
		}

		result = append(result, remoteHostAuth{Host: host, Auth: *remote.Auth})
	}

	return result, nil
}

// resolvePluginSource classifies a plugin source as local or remote and returns
// its normalized runtime metadata.
func resolvePluginSource(baseDir, rawSrc string, remotes []remoteHostAuth) (pluginSource, error) {
	src := strings.TrimSpace(rawSrc)
	if src == "" {
		return pluginSource{}, fmt.Errorf("plugin src is required")
	}

	if strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "http://") {
		canonicalURL, err := canonicalizeRemotePluginURL(src)
		if err != nil {
			return pluginSource{}, err
		}

		headers, err := resolveRemoteHeaders(canonicalURL, remotes)
		if err != nil {
			return pluginSource{}, err
		}

		return pluginSource{
			Kind:         pluginSourceKindRemote,
			Original:     src,
			DisplayName:  canonicalURL,
			CanonicalURL: canonicalURL,
			AuthMatchURL: canonicalURL,
			Headers:      headers,
		}, nil
	}

	if githubParts := githubPluginPattern.FindStringSubmatch(src); githubParts != nil {
		owner := githubParts[1]
		repo := githubParts[2]
		ref := githubParts[3]

		if !strings.HasPrefix(repo, "vdl-plugin-") {
			return pluginSource{}, fmt.Errorf(
				"GitHub plugin %q is invalid: repository name %q must start with %q",
				src,
				repo,
				"vdl-plugin-",
			)
		}

		canonicalURL := fmt.Sprintf(
			"https://raw.githubusercontent.com/%s/%s/%s/dist/index.js",
			owner,
			repo,
			ref,
		)
		authMatchURL := fmt.Sprintf("https://github.com/%s/%s/dist/index.js", owner, repo)

		headers, err := resolveRemoteHeaders(authMatchURL, remotes)
		if err != nil {
			return pluginSource{}, err
		}

		return pluginSource{
			Kind:         pluginSourceKindRemote,
			Original:     src,
			DisplayName:  src,
			CanonicalURL: canonicalURL,
			AuthMatchURL: authMatchURL,
			Headers:      headers,
		}, nil
	}

	if !isLocalPluginSource(src) {
		return pluginSource{}, fmt.Errorf(
			"plugin src %q is invalid: expected a local .js path, an HTTPS .js URL, or GitHub shorthand owner/repo@ref",
			src,
		)
	}

	localPath, err := resolveAbsolutePath(baseDir, src, ".js", "plugin src")
	if err != nil {
		return pluginSource{}, err
	}

	return pluginSource{
		Kind:        pluginSourceKindLocal,
		Original:    src,
		DisplayName: localPath,
		LocalPath:   localPath,
	}, nil
}

// canonicalizeRemotePluginURL validates a remote plugin URL and returns its
// canonical form.
func canonicalizeRemotePluginURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("plugin src %q is invalid: %w", rawURL, err)
	}

	if parsed.Scheme != "https" {
		if parsed.Scheme != "http" || !insecureHTTPAllowed() {
			return "", fmt.Errorf(
				"plugin src %q is invalid: only HTTPS URLs are allowed unless %s=true",
				rawURL,
				"VDL_INSECURE_ALLOW_HTTP",
			)
		}
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("plugin src %q is invalid: missing host", rawURL)
	}
	if !strings.HasSuffix(parsed.Path, ".js") {
		return "", fmt.Errorf("plugin src %q is invalid: remote plugins must point to a .js file", rawURL)
	}

	parsed.Fragment = ""
	return parsed.String(), nil
}

// resolveRemoteHeaders returns the HTTP headers that should be attached to a
// remote plugin request after matching the most specific configured remote.
func resolveRemoteHeaders(rawURL string, remotes []remoteHostAuth) (http.Header, error) {
	matchedRemote, ok := matchRemoteHost(rawURL, remotes)
	if !ok {
		return make(http.Header), nil
	}

	return buildAuthHeaders(matchedRemote)
}

// matchRemoteHost selects the most specific configured remote whose host prefix
// matches the target URL.
func matchRemoteHost(rawURL string, remotes []remoteHostAuth) (remoteHostAuth, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return remoteHostAuth{}, false
	}

	target := strings.ToLower(parsed.Host)
	pathPart := strings.Trim(strings.TrimSpace(parsed.EscapedPath()), "/")
	if pathPart != "" {
		target += "/" + strings.ToLower(pathPart)
	}

	bestLength := -1
	var best remoteHostAuth
	for _, remote := range remotes {
		if target == remote.Host || strings.HasPrefix(target, remote.Host+"/") {
			if len(remote.Host) > bestLength {
				best = remote
				bestLength = len(remote.Host)
			}
		}
	}

	if bestLength == -1 {
		return remoteHostAuth{}, false
	}

	return best, true
}

// buildAuthHeaders materializes HTTP headers for a matched remote
// authentication configuration.
func buildAuthHeaders(remote remoteHostAuth) (http.Header, error) {
	methods := 0
	headers := make(http.Header)

	if remote.Auth.Github != nil {
		methods++
		token, err := readRequiredEnv(remote.Auth.Github.GetTokenEnv(), remote.Host)
		if err != nil {
			return nil, err
		}
		headers.Set("Authorization", "Bearer "+token)
	}

	if remote.Auth.Header != nil {
		methods++
		name, err := readRequiredEnv(remote.Auth.Header.GetNameEnv(), remote.Host)
		if err != nil {
			return nil, err
		}
		value, err := readRequiredEnv(remote.Auth.Header.GetValueEnv(), remote.Host)
		if err != nil {
			return nil, err
		}
		headers.Set(name, value)
	}

	if remote.Auth.Bearer != nil {
		methods++
		token, err := readRequiredEnv(remote.Auth.Bearer.GetTokenEnv(), remote.Host)
		if err != nil {
			return nil, err
		}
		headers.Set("Authorization", "Bearer "+token)
	}

	if remote.Auth.Basic != nil {
		methods++
		user, err := readRequiredEnv(remote.Auth.Basic.GetUserEnv(), remote.Host)
		if err != nil {
			return nil, err
		}
		pass, err := readRequiredEnv(remote.Auth.Basic.GetPassEnv(), remote.Host)
		if err != nil {
			return nil, err
		}
		credentials := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
		headers.Set("Authorization", "Basic "+credentials)
	}

	if methods > 1 {
		return nil, fmt.Errorf("remote %q must define exactly one authentication method", remote.Host)
	}

	return headers, nil
}

// readRequiredEnv returns the value of a required environment variable and
// reports which remote requires it when it is missing.
func readRequiredEnv(name, remoteHost string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("remote %q contains an empty environment variable name", remoteHost)
	}

	value, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("remote %q requires environment variable %q", remoteHost, name)
	}

	return value, nil
}

// resolveAbsolutePath resolves a path relative to the config directory and
// enforces an optional file extension.
func resolveAbsolutePath(baseDir, rawPath, expectedExt, fieldName string) (string, error) {
	path := strings.TrimSpace(rawPath)
	if path == "" {
		return "", fmt.Errorf("%s is required", fieldName)
	}

	if expectedExt != "" && !strings.HasSuffix(path, expectedExt) {
		return "", fmt.Errorf("%s %q is invalid: expected a %s file", fieldName, rawPath, expectedExt)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s %q: %w", fieldName, rawPath, err)
	}

	return absPath, nil
}

// normalizeRemoteHost strips optional schemes and separators so remote hosts
// can be compared consistently against request URLs.
func normalizeRemoteHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.Trim(host, "/")
	return strings.ToLower(host)
}

// isLocalPluginSource reports whether src uses the local-path form supported by
// the configuration schema.
func isLocalPluginSource(src string) bool {
	return filepath.IsAbs(src) || strings.HasPrefix(src, ".")
}

// cloneStringMap returns a shallow copy of values and always returns a
// non-nil map.
func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	return maps.Clone(values)
}

// insecureHTTPAllowed reports whether insecure HTTP plugin URLs are enabled for
// local development and bootstrapping scenarios.
func insecureHTTPAllowed() bool {
	value, ok := os.LookupEnv("VDL_INSECURE_ALLOW_HTTP")
	if !ok {
		return false
	}

	value = strings.TrimSpace(strings.ToLower(value))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
