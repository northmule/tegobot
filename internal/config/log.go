package config

type Log struct {
	Level   string `mapstructure:"Level"`
	FlePath string `mapstructure:"FilePath"`
}
