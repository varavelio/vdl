package openapi

type Spec struct {
	OpenAPI    string                `json:"openapi"`
	Info       Info                  `json:"info"`
	Servers    []Server              `json:"servers,omitempty"`
	Security   []map[string][]string `json:"security,omitempty"`
	Tags       []Tag                 `json:"tags,omitempty"`
	Paths      Paths                 `json:"paths,omitempty"`
	Components Components            `json:"components,omitzero"`
}

type Info struct {
	Title       string      `json:"title,omitzero"`
	Version     string      `json:"version,omitzero"`
	Description string      `json:"description,omitzero"`
	Contact     InfoContact `json:"contact,omitzero"`
	License     InfoLicense `json:"license,omitzero"`
}

type InfoContact struct {
	Name  string `json:"name,omitzero"`
	Email string `json:"email,omitzero"`
}

type InfoLicense struct {
	Name string `json:"name,omitzero"`
}

type Server struct {
	URL string `json:"url"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitzero"`
}

type Paths map[string]any

type Components struct {
	SecuritySchemes map[string]any `json:"securitySchemes,omitempty"`
	Schemas         map[string]any `json:"schemas,omitempty"`
	RequestBodies   map[string]any `json:"requestBodies,omitempty"`
	Responses       map[string]any `json:"responses,omitempty"`
}
