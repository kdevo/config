package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kdevo/config"
)

type JSONProvider[T config.Config] struct {
	name string
	path string
}

func JSON[T config.Config](path string) *JSONProvider[T] {
	return &JSONProvider[T]{
		name: "json",
		path: path,
	}
}

func (p *JSONProvider[T]) WithName(name string) *JSONProvider[T] {
	p.name = name
	return p
}

func (p *JSONProvider[T]) Name() string {
	return p.name
}

func (p *JSONProvider[T]) Config() (T, error) {
	var target T
	r, err := os.Open(p.path)
	if err != nil {
		return target, fmt.Errorf("could not open file: %w", err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return target, fmt.Errorf("could not read file: %w", err)
	}
	err = json.Unmarshal(data, &target)
	if err != nil {
		return target, fmt.Errorf("could not unmarshal: %w", err)
	}
	return target, nil
}
