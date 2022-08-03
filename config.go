package config

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ConfigProvider[T Config] interface {
	Config() (T, error)
	Name() string
}

type FieldAwareConfigProvider[T Config] interface {
	ConfigProvider[T]
	Inject(fields Fields)
}

type Config interface {
	Validate() error
}

// Fields is a mapping from field name to value.
type Fields map[string]any

// Loader is a simple config loader which makes use of (un)marshalling to a generic map.
// It therewith indirects any tedious reflection access.
type Loader[T Config] struct {
	providers      []ConfigProvider[T]
	providerErrors ProviderErrors
}

func From[T Config](cs ...ConfigProvider[T]) *Loader[T] {
	return &Loader[T]{
		providers:      cs,
		providerErrors: make(map[string]error, len(cs)),
	}
}

func (l *Loader[T]) WithDefaults(cs ...ConfigProvider[T]) *Loader[T] {
	l.providers = append(l.providers, cs...)
	return l
}

func (l *Loader[T]) Resolve() (T, error) {
	var target T
	// If it's a pointer, we can't safely ensure that the struct is non-nil.
	// However, we need the struct to be initialized (just empty values) to be able to iterate through fields.
	if reflect.ValueOf(target).Kind() == reflect.Pointer {
		return target, fmt.Errorf("provided config T must not be a pointer")
	}
	fields, err := ToFields(target)
	if err != nil {
		return target, err
	}
	resolved := make(Fields, len(fields))
	for _, p := range l.providers {
		// Inject fields if supported by provider:
		if fieldAwareProvider, ok := p.(FieldAwareConfigProvider[T]); ok {
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
		providerConfig, err := ToFields(cfg)
		if err != nil {
			return target, err
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
			// TODO(kdevo): T must be the same for each provider, so we probably don't need this check anymore:
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
	target, err = ToConfig[T](resolved)
	if err != nil {
		return target, err
	}
	return target, target.Validate()
}

func (p *Loader[T]) ProviderErrors() ProviderErrors {
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

func ToFields(v Config) (Fields, error) {
	d, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("can not convert to JSON: %w", err)
	}
	var res Fields
	err = json.Unmarshal(d, &res)
	if err != nil {
		return nil, fmt.Errorf("can not convert to generic Fields map: %w", err)
	}
	return res, nil
}

func ToConfig[T Config](m Fields) (T, error) {
	var target T
	d, err := json.Marshal(m)
	if err != nil {
		return target, fmt.Errorf("can not convert to JSON: %w", err)
	}
	err = json.Unmarshal(d, &target)
	if err != nil {
		return target, fmt.Errorf("can not convert to target: %w", err)
	}
	return target, nil
}
