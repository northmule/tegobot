package config

type Akismet struct {
	ApiKey  string `mapstructure:"ApiKey"`
	SiteURL string `mapstructure:"SiteURL"`
}
