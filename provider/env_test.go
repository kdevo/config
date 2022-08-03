package provider_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/kdevo/config"
	"github.com/kdevo/config/provider"
)

type Config struct {
	HTTPTimeout duration `json:"HTTPTimeout"`
	RepoOwner   string   `json:"owner"`
	Token       token    `json:"MyToken"`
}

type token string
type duration time.Duration

func (d *duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	td, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = duration(td)
	return nil
}

func (c Config) Validate() error {
	return nil
}

func TestConfig(t *testing.T) {
	testCases := []struct {
		name       string
		env        EnvMap
		wantConfig Config
	}{
		{
			name: "load from env with camelCase",
			env: EnvMap{
				"HTTP_TIMEOUT": "1s",
				"OWNER":        "kdevo",
				"MY_TOKEN":     "test",
			},
			wantConfig: Config{
				HTTPTimeout: duration(1 * time.Second),
				RepoOwner:   "kdevo",
				Token:       "test",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := provider.Environment[Config]().WithEnvLookupFunc(tc.env.GetEnv)
			fields, _ := config.ToFields(Config{})
			c.Inject(fields)
			cfg, err := c.Config()
			if err != nil {
				t.Errorf("got unwanted error: %v", err)
			}
			if !reflect.DeepEqual(tc.wantConfig, cfg) {
				t.Errorf("unexpected result:\n  want=%v\n   got=%v", tc.wantConfig, cfg)
			}
		})
	}
}

type EnvMap map[string]string

func (e EnvMap) GetEnv(name string) (string, bool) {
	v, ok := e[name]
	return v, ok
}

func TestCamelToUpperCase(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{
			input: "a",
			want:  "A",
		},
		{
			input: "HELLO",
			want:  "HELLO",
		},
		{
			input: "HelloWorld",
			want:  "HELLO_WORLD",
		},
		{
			input: "HelloW",
			want:  "HELLO_W",
		},
		{
			input: "HTTPConnectionTimeout",
			want:  "HTTP_CONNECTION_TIMEOUT",
		},
		{
			input: "httpConnections",
			want:  "HTTP_CONNECTIONS",
		},
		{
			input: "usingTheGNUAsAcronym",
			want:  "USING_THE_GNU_AS_ACRONYM",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			got := provider.CamelToUpperCase(tc.input)
			if tc.want != got {
				t.Errorf("   want %q,\nbut got %q", tc.want, got)
			}
		})
	}
}
