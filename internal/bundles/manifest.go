package bundles

import (
	_ "embed"
	"encoding/json"
	"sort"
)

//go:embed manifest.json
var manifestBytes []byte

type Manifest struct {
	SchemaVersion int                 `json:"schemaVersion"`
	Platforms     map[string]Platform `json:"platforms"`
}

type Platform map[string]Component

type Component struct {
	Version string `json:"version"`
	Source  string `json:"source"`
	Notes   string `json:"notes"`
}

type NamedComponent struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Source   string `json:"source"`
	Notes    string `json:"notes"`
	Platform string `json:"platform"`
}

func MustLoad() Manifest {
	var manifest Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		panic(err)
	}
	return manifest
}

func ComponentsForPlatform(platform string) []NamedComponent {
	manifest := MustLoad()
	target, ok := manifest.Platforms[platform]
	if !ok {
		return nil
	}

	components := make([]NamedComponent, 0, len(target))
	for name, component := range target {
		components = append(components, NamedComponent{
			Name:     name,
			Version:  component.Version,
			Source:   component.Source,
			Notes:    component.Notes,
			Platform: platform,
		})
	}
	sort.Slice(components, func(i, j int) bool { return components[i].Name < components[j].Name })
	return components
}
