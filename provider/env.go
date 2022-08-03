package provider

import (
	"encoding/json"
	"os"
	"strings"
	"unicode"

	"github.com/kdevo/config"
)

type caseConverterFunc func(string) string
type envLookupFunc func(string) (string, bool)

type EnvProvider[T config.Config] struct {
	name          string
	fields        config.Fields
	caseConverter caseConverterFunc
	env           envLookupFunc
}

func Environment[T config.Config]() *EnvProvider[T] {
	return &EnvProvider[T]{
		name:          "environment",
		caseConverter: CamelToUpperCase,
		env:           os.LookupEnv,
	}
}

func (p *EnvProvider[T]) Name() string {
	return p.name
}

func (p *EnvProvider[T]) WithName(n string) *EnvProvider[T] {
	p.name = n
	return p
}

func (p *EnvProvider[T]) WithCaseConverter(cc caseConverterFunc) *EnvProvider[T] {
	p.caseConverter = cc
	return p
}

func (p *EnvProvider[T]) WithEnvLookupFunc(e envLookupFunc) *EnvProvider[T] {
	p.env = e
	return p
}

func (p *EnvProvider[T]) Inject(fields config.Fields) {
	p.fields = fields
}

func (p *EnvProvider[T]) Config() (T, error) {
	m := make(map[string]any)
	for f := range p.fields {
		if v, ok := p.env(p.caseConverter(f)); ok {
			m[f] = v
		}
	}
	d, err := json.Marshal(m)
	var target T
	if err != nil {
		return target, err
	}
	err = json.Unmarshal(d, &target)
	return target, err
}

func CamelToUpperCase(field string) string {
	// Example: Let's say we have field HTTPConnectionTimeout and want HTTP_CONNECTION_TIMEOUT:
	// 1. Collect the breakpoints where to split the camelCase words
	// 2. Convert words to upper case, concatenate
	var breakpoints []int
	lastUpper := false
	acronym := false
	for i, r := range field {
		currentUpper := unicode.IsUpper(r)
		if i > 0 {
			switch {
			case !acronym && currentUpper && !lastUpper:
				// basic case: change to upper, e.g.:
				// aNewWorld
				//  ^
				breakpoints = append(breakpoints, i)
			case !acronym && currentUpper && lastUpper:
				// special case: acronym start, e.g.:
				// GNUIsNotUnix
				//  ^
				acronym = true
			case acronym && !currentUpper && lastUpper:
				// special case: acronym ended, e.g.:
				// GNUIsNotUnix
				//     ^
				breakpoints = append(breakpoints, i-1)
				acronym = false
			}
		}
		lastUpper = currentUpper
	}

	var sb strings.Builder
	sb.Grow(len(field) + len(breakpoints))
	last := 0
	for _, bp := range breakpoints {
		sb.WriteString(strings.ToUpper(field[last:bp]))
		sb.WriteByte('_')
		last = bp
	}
	sb.WriteString(strings.ToUpper(field[last:]))
	return sb.String()
}

func NoOPCaseConverter(field string) string {
	return field
}
