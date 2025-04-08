package config

import (
	"github.com/spf13/viper"
)

// Config конфигурация
type Config struct {
	v     *viper.Viper
	value *Value
}

// Spam Раздел по спам проверкам
type Spam struct {
	Akismet Akismet `mapstructure:"Akismet"`
	Common  Common  `mapstructure:"Common"`
}

// Value Вся конфигурация
type Value struct {
	Telegram Telegram `mapstructure:"Telegram"`
	Spam     Spam     `mapstructure:"Spam"`
	Log      Log      `mapstructure:"Log"`
}

// ErrorCfg ошибка конфигурации
type ErrorCfg error

// NewConfig конструктор
func NewConfig() (*Config, error) {
	var err error
	instance := new(Config)
	instance.v = viper.New()
	instance.value = new(Value)

	err = instance.init()
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (c *Config) init() error {
	var err error
	c.v.AddConfigPath(".")
	c.v.SetConfigName("app_config")
	c.v.SetConfigType("yaml")
	err = c.v.ReadInConfig()
	if err != nil {
		return ErrorCfg(err)
	}

	err = c.v.Unmarshal(c.value)
	if err != nil {
		return ErrorCfg(err)
	}

	return nil
}

func (c *Config) Value() *Value {
	return c.value
}
