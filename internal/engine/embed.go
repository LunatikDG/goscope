package engine

import (
	"embed"
	"fmt"
)

//go:embed scenes/*.yaml
var scenesFS embed.FS

// LoadScene loads a built-in scene by file name without extension,
// e.g. LoadScene("workerpool") reads scenes/workerpool.yaml.
func LoadScene(name string) (Scene, error) {
	data, err := scenesFS.ReadFile("scenes/" + name + ".yaml")
	if err != nil {
		return Scene{}, fmt.Errorf("сцена %q не найдена: %w", name, err)
	}
	return ParseScene(data)
}
