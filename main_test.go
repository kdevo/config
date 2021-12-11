package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kdevo/gocfg/pkg/config"
	"github.com/kdevo/gocfg/pkg/provider"
)

// Define your configuration struct:
type Config struct {
	RepoOwner   string
	RepoName    string
	URL         string
	HTTPTimeout time.Duration
}

// A config is complete when it can be validated.
// By optionally implementing the Config interface's Validate, we can detect errors.
func (c *Config) Validate() error {
	var errors config.Errors
	if !strings.HasPrefix(c.URL, "http") {
		// Capture errors on a per-field basis by naming the errors accordingly.
		// This way, the config loader knows which fields are valid.
		errors.Add(config.Err("URL", c.URL, "must be a valid URL (starting with http)"))
	}
	if c.HTTPTimeout < 1*time.Second {
		errors.Add(config.Err("HTTPTimeout", c.HTTPTimeout, "must be >= 1s"))
	}
	return errors.AsError() // returns nil if we did not add any error
}

// Config providers follow below signature and return a config and eventually an error.
// There is nothing that prevents the config from being a provider for itself:
func (c *Config) Config() (interface{}, error) {
	return c, c.Validate() // makes some things easier later on
}

// Name the config provider accordingly to make it easily identifiable later on.
func (c *Config) Name() string {
	return "Static"
}

func TestMain(t *testing.T) {
	// get config from 'JSON' first, fallback to 'Static' defaults otherwise:
	loader := config.From(provider.JSON("config.json")).
		WithDefaults(&Config{
			RepoOwner: "kdevo",
			RepoName:  "gocfg",
		}) // works because we've implemented the config provider interface above
	var cfg Config
	err := loader.Resolve(&cfg) // calls Validate if implemented
	if err != nil {
		// examine errors by providers to find out the cause:
		fmt.Printf("providers errors: %v\n", loader.ProviderErrors())
	}
	fmt.Printf("got config: %v\n", cfg)
}
