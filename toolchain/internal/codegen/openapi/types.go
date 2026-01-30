package openapi

type Spec struct {
	OpenAPI    string                `json:"openapi" yaml:"openapi"`
	Info       Info                  `json:"info" yaml:"info"`
	Servers    []Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Security   []map[string][]string `json:"security,omitempty" yaml:"security,omitempty"`
	Tags       []Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	Paths      Paths                 `json:"paths,omitempty" yaml:"paths,omitempty"`
	Components Components            `json:"components,omitzero" yaml:"components,omitempty"`
}

type Info struct {
	Title       string      `json:"title,omitzero" yaml:"title,omitempty"`
	Version     string      `json:"version,omitzero" yaml:"version,omitempty"`
	Description string      `json:"description,omitzero" yaml:"description,omitempty"`
	Contact     InfoContact `json:"contact,omitzero" yaml:"contact,omitempty"`
	License     InfoLicense `json:"license,omitzero" yaml:"license,omitempty"`
}

type InfoContact struct {
	Name  string `json:"name,omitzero" yaml:"name,omitempty"`
	Email string `json:"email,omitzero" yaml:"email,omitempty"`
}

type InfoLicense struct {
	Name string `json:"name,omitzero" yaml:"name,omitempty"`
}

type Server struct {
	URL string `json:"url" yaml:"url"`
}

type Tag struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitzero" yaml:"description,omitempty"`
}

type Paths map[string]any

type Components struct {
	SecuritySchemes map[string]any `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Schemas         map[string]any `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	RequestBodies   map[string]any `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Responses       map[string]any `json:"responses,omitempty" yaml:"responses,omitempty"`
}
