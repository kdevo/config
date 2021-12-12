package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type JSONProvider struct {
	name string
	path string
}

func JSON(path string) *JSONProvider {
	return &JSONProvider{
		name: "JSON",
		path: path,
	}
}

func (p *JSONProvider) WithName(name string) *JSONProvider {
	p.name = name
	return p
}

func (p *JSONProvider) Name() string {
	return p.name
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
	var config interface{} // TODO(kdevo): use target type instead
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal: %w", err)
	}
	return &config, nil
}
