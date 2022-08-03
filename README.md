## config

> :warning: Work in progress!

A simple [multi-layered config loader](#rules) for Go.

Made for smaller projects. No external dependencies.

### Installation

```
go get -u github.com/kdevo/config
```

### Example

```go
// Define your configuration struct:
type Config struct {
	RepoOwner   string
	RepoName    string
	URL         string
	HTTPTimeout time.Duration
}

// A config is complete when it can be validated.
// By optionally implementing the Config interface's Validate, we can detect errors.
func (c Config) Validate() error {
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
func (c Config) Config() (Config, error) {
	return c, c.Validate() // makes some things easier later on
}

// Name the config provider accordingly to make it easily identifiable later on.
func (c Config) Name() string {
	return "Static"
}

func main() {
	// get config from 'JSON' first, fallback to 'Static' defaults otherwise.
	// we can have any layer of chained config providers by using the builder funcs:
	loader := config.From(provider.JSON("config.json")).
		WithDefaults(Config{
			RepoOwner: "kdevo",
			RepoName:  "config",
		}) // works because we've implemented the config provider interface above
	err := loader.Resolve() // calls Validate
	if err != nil {
		// examine errors by providers to find out the cause:
		fmt.Printf("providers errors: %v\n", loader.ProviderErrors())
	}
	fmt.Printf("got config: %v\n", cfg)
}
```

Please also take a look at the [config loader test](./config_test.go).

## Loading Rules <a id="rules"></a>

1. Earlier config providers take precedence.
2. Config structs must be marshable.
3. [Empty/ommitted fields](https://pkg.go.dev/encoding/json) are skipped.
4. Fields with errors are skipped (see below).

### Provider Errors 

1. Returning `config.Errors` with the specified field names will skip the field.
2. Field names must follow rules of [encoding/json](https://pkg.go.dev/encoding/json).
3. Returning regular errors will skip the entire provider (e.g. if the config is corrupt).

Tips:
1. Skip fields that couldn't be provided by using `config.Errors` (e.g. when a value has an invalid format).
2. Use descriptive field names that are unlikely to change.
3. Return a regular error if we can't provide anything (e.g. Unmarshal error for [JSON provider](./pkg/provider/json.go)).
