package provider

import "github.com/kdevo/config"

type FunctionalProvider[T config.Config] struct {
	name         string
	providerFunc func() (T, error)
}

func Function[T config.Config](name string, fn func() (T, error)) *FunctionalProvider[T] {
	if name == "" {
		name = "function"
	}
	return &FunctionalProvider[T]{
		name:         name,
		providerFunc: fn,
	}
}

func (p *FunctionalProvider[T]) WithName(name string) *FunctionalProvider[T] {
	p.name = name
	return p
}

func (p *FunctionalProvider[T]) Config() (T, error) {
	return p.providerFunc()
}

func (p *FunctionalProvider[T]) Name() string {
	return p.name
}
