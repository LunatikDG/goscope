package engine

import (
	"os"
	"reflect"
	"testing"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("не удалось прочитать фикстуру %s: %v", name, err)
	}
	return data
}

func TestParseSceneValid(t *testing.T) {
	scene, err := ParseScene(readFixture(t, "valid.yaml"))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	want := Scene{
		Name: "Fixture: minimal",
		Steps: []Step{
			{Event: Spawn, Goroutine: 5, Label: "solo"},
			{Event: Done, Goroutine: 5},
		},
	}
	if !reflect.DeepEqual(scene, want) {
		t.Errorf("сцена = %+v, ожидалось %+v", scene, want)
	}
}

func TestParseSceneInvalid(t *testing.T) {
	tests := []string{
		"invalid_unknown_event.yaml",
		"invalid_missing_goroutine.yaml",
		"invalid_empty_steps.yaml",
		"invalid_missing_name.yaml",
		"invalid_negative_chan.yaml",
		"invalid_malformed.yaml",
	}

	for _, fixture := range tests {
		t.Run(fixture, func(t *testing.T) {
			_, err := ParseScene(readFixture(t, fixture))
			if err == nil {
				t.Fatalf("ожидалась ошибка для %s, получили nil", fixture)
			}
		})
	}
}

// LoadScene("workerpool") должен разворачиваться в те же кадры, что и
// программный конструктор WorkerPool(3) — YAML-версия его точная копия.
func TestLoadSceneWorkerPoolMatchesBuilder(t *testing.T) {
	fromYAML, err := LoadScene("workerpool")
	if err != nil {
		t.Fatalf("LoadScene: %v", err)
	}
	fromBuilder := WorkerPool(3)

	if !reflect.DeepEqual(fromYAML.Steps, fromBuilder.Steps) {
		t.Fatalf("шаги YAML-сцены расходятся с WorkerPool(3):\nYAML:    %+v\nbuilder: %+v", fromYAML.Steps, fromBuilder.Steps)
	}
	if !reflect.DeepEqual(fromYAML.Frames(), fromBuilder.Frames()) {
		t.Fatal("кадры YAML-сцены расходятся с кадрами WorkerPool(3)")
	}
}

func TestLoadSceneUnknownName(t *testing.T) {
	if _, err := LoadScene("no-such-scene"); err == nil {
		t.Fatal("ожидалась ошибка для несуществующей сцены")
	}
}
