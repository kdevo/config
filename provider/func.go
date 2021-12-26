package provider

type FunctionalProvider struct {
	name         string
	providerFunc func() (interface{}, error)
}

func Function(name string, fn func() (interface{}, error)) *FunctionalProvider {
	if name == "" {
		name = "function"
	}
	return &FunctionalProvider{
		name:         name,
		providerFunc: fn,
	}
}

func (p *FunctionalProvider) WithName(name string) *FunctionalProvider {
	p.name = name
	return p
}

func (p *FunctionalProvider) Config() (interface{}, error) {
	return p.providerFunc()
}

func (p *FunctionalProvider) Name() string {
	return p.name
}
