package config

import (
	"encoding/json"
	"fmt"
)

type ConfigProvider interface {
	Config() (interface{}, error)
	Name() string
}

type FieldAwareConfigProvider interface {
	ConfigProvider
	Inject(fields Fields)
}

type Config interface {
	Validate() error
}

// Fields is a mapping from field name to value.
type Fields map[string]interface{}

// Loader is a simple config loader which makes use of (un)marshalling to a generic map.
// It therewith indirects any tedious reflection access.
type Loader struct {
	providers      []ConfigProvider
	providerErrors ProviderErrors
}

func From(cs ...ConfigProvider) *Loader {
	return &Loader{
		providers:      cs,
		providerErrors: make(map[string]error, len(cs)),
	}
}

func (l *Loader) WithDefaults(cs ...ConfigProvider) *Loader {
	l.providers = append(l.providers, cs...)
	return l
}

func (l *Loader) Resolve(target interface{}) error {
	fields, err := ToFields(target)
	if err != nil {
		return err
	}
	resolved := make(map[string]interface{}, len(fields))
	for _, p := range l.providers {
		// Inject fields if supported by provider:
		if fieldAwareProvider, ok := p.(FieldAwareConfigProvider); ok {
			fieldAwareProvider.Inject(fields)
		}
		// Try to get the config, skip if a not recoverable err is returned:
		var configErrors *Errors
		cfg, err := p.Config()
		if err != nil {
			l.providerErrors[p.Name()] = err
			errs, ok := err.(*Errors)
			if !ok {
				continue
			}
			configErrors = errs
		}
		// Capture maps directly if returned by provider:
		var providerConfig map[string]interface{}
		if m, ok := cfg.(map[string]interface{}); ok {
			providerConfig = m
		} else {
			providerConfig, err = ToFields(cfg)
			if err != nil {
				return err
			}
		}
		for field := range fields {
			// skip if already resolved:
			if _, ok := resolved[field]; ok {
				continue
			}
			// skip if there is an error with the field:
			if configErrors != nil && configErrors.HasField(field) {
				continue
			}
			// skip if field is not available:
			val, ok := providerConfig[field]
			if !ok {
				continue
			}
			// skip if field is empty:
			if isEmpty := val == fields[field]; isEmpty {
				continue
			}
			resolved[field] = val
		}
		if allResolved := len(resolved) == len(fields); allResolved {
			break
		}
	}
	err = ToConfig(resolved, target)
	if err != nil {
		return err
	}
	if targetConfig, ok := target.(Config); ok {
		err = targetConfig.Validate()
	}
	return err
}

func (p *Loader) ProviderErrors() ProviderErrors {
	return p.providerErrors
}

type ProviderErrors map[string]error

func (e ProviderErrors) Providers() []string {
	providers := make([]string, 0, len(e))
	for p := range e {
		providers = append(providers, p)
	}
	return providers
}

func ToFields(v interface{}) (Fields, error) {
	d, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("can not convert to JSON: %w", err)
	}
	res := map[string]interface{}{}
	err = json.Unmarshal(d, &res)
	if err != nil {
		return nil, fmt.Errorf("can not convert to generic map: %w", err)
	}
	return res, nil
}

func ToConfig(m map[string]interface{}, target interface{}) error {
	d, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("can not convert to JSON: %w", err)
	}
	err = json.Unmarshal(d, target)
	if err != nil {
		return fmt.Errorf("can not convert to target: %w", err)
	}
	return nil
}
