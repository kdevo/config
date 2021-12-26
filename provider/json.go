package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type JSONProvider struct {
	name   string
	path   string
	target interface{}
}

func JSON(path string) *JSONProvider {
	return &JSONProvider{
		name:   "json",
		path:   path,
		target: nil,
	}
}

func (p *JSONProvider) WithName(name string) *JSONProvider {
	p.name = name
	return p
}

func (p *JSONProvider) Name() string {
	return p.name
}

func (p *JSONProvider) WithTarget(typ interface{}) *JSONProvider {
	p.target = typ
	return p
}

func (p *JSONProvider) Config() (interface{}, error) {
	r, err := os.Open(p.path)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	target := p.target
	err = json.Unmarshal(data, &target)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal: %w", err)
	}
	return target, nil
}
