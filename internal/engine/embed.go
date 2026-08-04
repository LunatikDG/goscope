package engine

import (
	"embed"
	"fmt"
)

//go:embed scenes/*.yaml
var scenesFS embed.FS

// LoadScene загружает встроенную сцену по имени файла без расширения,
// напр. LoadScene("workerpool") читает scenes/workerpool.yaml.
func LoadScene(name string) (Scene, error) {
	data, err := scenesFS.ReadFile("scenes/" + name + ".yaml")
	if err != nil {
		return Scene{}, fmt.Errorf("сцена %q не найдена: %w", name, err)
	}
	return ParseScene(data)
}
