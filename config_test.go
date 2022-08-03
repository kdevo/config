package config_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kdevo/config"
	"github.com/kdevo/config/provider"
)

type Config struct {
	RepoOwner   string `json:"owner"`
	RepoName    string `json:"name"`
	URL         string
	HTTPTimeout time.Duration
}

func (c Config) Config() (Config, error) {
	return c, c.Validate()
}

func (c Config) Name() string {
	return "Config"
}

func (c Config) Validate() error {
	var errors config.Errors
	if c.RepoOwner == "" {
		errors.Add(config.Err("RepoOwner", c.RepoOwner, "must not be empty"))
	}
	if c.RepoName == "" {
		errors.Add(config.Err("RepoName", c.RepoName, "must not be empty"))
	}
	if !strings.HasPrefix(c.URL, "http") {
		errors.Add(config.Err("URL", c.URL, "must be a valid URL (starting with http)"))
	}
	if c.HTTPTimeout < 1*time.Second {
		errors.Add(config.Err("HTTPTimeout", c.HTTPTimeout, "must be >= 1s"))
	}
	return errors.AsError()

}

func Test_Loader(t *testing.T) {
	testCases := []struct {
		name        string
		loader      *config.Loader[Config]
		want        Config
		wantIsValid bool
	}{
		{
			name:   "load without defaults",
			loader: config.From[Config](Config{RepoOwner: "kdevo"}),
			want: Config{
				RepoOwner: "kdevo",
			},
			wantIsValid: false,
		},
		{
			name: "load missing with defaults",
			loader: config.From[Config](Config{RepoOwner: "kdevo"}).
				WithDefaults(Config{RepoName: "osprey-delight"}),
			want:        Config{RepoOwner: "kdevo", RepoName: "osprey-delight"},
			wantIsValid: false,
		},
		{
			name: "load with defaults without default overriding higher priority",
			loader: config.From[Config](Config{RepoOwner: "kdevo"}).
				WithDefaults(Config{RepoOwner: "hugo-mods"}),
			want: Config{
				RepoOwner: "kdevo",
			},
			wantIsValid: false,
		},
		{
			name: "load all with multi-layered defaults",
			loader: config.From[Config](Config{
				RepoOwner: "hugo-mods",
				RepoName:  "discussions",
			}).WithDefaults(Config{
				RepoOwner:   "test",
				HTTPTimeout: 1 * time.Microsecond,
			}).WithDefaults(Config{
				HTTPTimeout: 2 * time.Second,
				URL:         "https://hugo-mods.github.io/sitemap.xml",
			}),
			want: Config{
				RepoOwner:   "hugo-mods",
				RepoName:    "discussions",
				HTTPTimeout: 2 * time.Second,
				URL:         "https://hugo-mods.github.io/sitemap.xml",
			},
			wantIsValid: true,
		},
		{
			name: "load all with multi-layered defaults via different providers",
			loader: config.From[Config](provider.Function("config", func() (Config, error) {
				return Config{RepoName: "discussions"}, nil
			},
			)).WithDefaults(
				provider.Environment[Config]().WithEnvLookupFunc(func(n string) (string, bool) {
					v, ok := map[string]string{"OWNER": "hugo-mods"}[n] // simulate env variable
					return v, ok
				}),
				provider.JSON[Config]("testdata/config.json"),
				Config{
					HTTPTimeout: 11 * time.Second,
				},
				Config{
					HTTPTimeout: 1 * time.Second,
					URL:         "https://hugo-mods.github.io/sitemap.xml",
				},
			),
			want: Config{
				RepoOwner:   "hugo-mods",                               // from provider.Environment
				RepoName:    "discussions",                             // from provider.Function
				URL:         "https://hugo-mods.github.io/sitemap.xml", // from first Config
				HTTPTimeout: 11 * time.Second,                          // from first Config
			},
			wantIsValid: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.loader.Resolve()
			if isValid := err == nil; isValid != tc.wantIsValid {
				t.Errorf("want isValid = %t, but got = %t: %v", tc.wantIsValid, isValid, err)
			}
			if !reflect.DeepEqual(tc.want, got) {
				t.Errorf("unexpected result:\n  want=%s\n   got=%s", tc.want, got)
			}
		})
	}
}
