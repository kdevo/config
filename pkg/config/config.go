package config

type ConfigProvider interface {
	Config() (interface{}, error)
	Name() string
}

type Config interface {
	Validate() error
}
