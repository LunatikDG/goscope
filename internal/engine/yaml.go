package engine

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// yamlScene — the schema of a scene's YAML file.
type yamlScene struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Steps       []yamlStep `yaml:"steps"`
}

// yamlStep — the schema of one step; Goroutine is a pointer so we can tell an
// omitted field apart from an explicit zero (goroutine id 0 is a valid actor).
type yamlStep struct {
	Goroutine *int   `yaml:"goroutine"`
	Event     string `yaml:"event"`
	Label     string `yaml:"label"`
	Chan      int    `yaml:"chan"`
}

// ParseScene parses a scene's YAML description and validates it.
func ParseScene(data []byte) (Scene, error) {
	var raw yamlScene
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Scene{}, fmt.Errorf("разбор yaml: %w", err)
	}
	return validateScene(raw)
}

func validateScene(raw yamlScene) (Scene, error) {
	if raw.Name == "" {
		return Scene{}, errors.New("сцена: поле name обязательно")
	}
	if len(raw.Steps) == 0 {
		return Scene{}, fmt.Errorf("сцена %q: должен быть хотя бы один шаг", raw.Name)
	}

	steps := make([]Step, len(raw.Steps))
	for i, rs := range raw.Steps {
		event, err := ParseEventType(rs.Event)
		if err != nil {
			return Scene{}, fmt.Errorf("сцена %q: шаг %d: %w", raw.Name, i, err)
		}
		if rs.Goroutine == nil {
			return Scene{}, fmt.Errorf("сцена %q: шаг %d: поле goroutine обязательно", raw.Name, i)
		}
		if *rs.Goroutine < 0 {
			return Scene{}, fmt.Errorf("сцена %q: шаг %d: goroutine не может быть отрицательным", raw.Name, i)
		}
		if rs.Chan < 0 {
			return Scene{}, fmt.Errorf("сцена %q: шаг %d: chan не может быть отрицательным", raw.Name, i)
		}

		steps[i] = Step{
			Label:     rs.Label,
			Event:     event,
			Goroutine: *rs.Goroutine,
			Chan:      rs.Chan,
		}
	}

	return Scene{Name: raw.Name, Description: raw.Description, Steps: steps}, nil
}
