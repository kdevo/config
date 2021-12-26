package provider

import (
	"os"
	"strings"
	"unicode"

	"github.com/kdevo/config"
)

type caseConverterFunc func(string) string
type envLookupFunc func(string) (string, bool)

type EnvProvider struct {
	name          string
	fields        config.Fields
	caseConverter caseConverterFunc
	envFunc       envLookupFunc
}

func Environment() *EnvProvider {
	return &EnvProvider{
		name:          "environment",
		caseConverter: CamelToUpperCase,
		envFunc:       os.LookupEnv,
	}
}

func (p *EnvProvider) Name() string {
	return p.name
}

func (p *EnvProvider) WithName(n string) *EnvProvider {
	p.name = n
	return p
}

func (p *EnvProvider) WithCaseConverter(cc caseConverterFunc) *EnvProvider {
	p.caseConverter = cc
	return p
}

func (p *EnvProvider) WithEnvLookupFunc(e envLookupFunc) *EnvProvider {
	p.envFunc = e
	return p
}

func (p *EnvProvider) Inject(fields config.Fields) {
	p.fields = fields
}

func (p *EnvProvider) Config() (interface{}, error) {
	m := make(map[string]interface{})
	for f := range p.fields {
		if v, ok := p.envFunc(p.caseConverter(f)); ok {
			m[f] = v
		}
	}
	return m, nil
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
