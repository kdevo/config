package provider

type DynamicProvider struct {
	name         string
	providerFunc func() (interface{}, error)
}

func Dynamic(fn func() (interface{}, error)) *DynamicProvider {
	return &DynamicProvider{
		name:         "Dynamic",
		providerFunc: fn,
	}
}

func (p *DynamicProvider) WithName(name string) *DynamicProvider {
	p.name = name
	return p
}

func (p *DynamicProvider) Config() (interface{}, error) {
	return p.providerFunc()
}

func (p *DynamicProvider) Name() string {
	return p.name
}
