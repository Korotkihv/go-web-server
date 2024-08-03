package config_loader

import (
	"github.com/spf13/viper"
)

type Route struct {
	Name      string `mapstructure:"name"`
	Context   string `mapstructure:"context"`
	Target    string `mapstructure:"target"`
	Versioned bool   `mapstructure:"versioned"`
}

type AggregatedRoute struct {
	Name    string        `mapstructure:"name"`
	Context string        `mapstructure:"context"`
	Targets []TargetRoute `mapstructure:"targets"`
}

type ChainedRoute struct {
	Name    string        `mapstructure:"name"`
	Context string        `mapstructure:"context"`
	Targets []TargetRoute `mapstructure:"targets"`
}

type TargetRoute struct {
	Target    string `mapstructure:"target"`
	Versioned bool   `mapstructure:"versioned"`
}

type GatewayConfig struct {
	ListenAddr       string            `mapstructure:"listenAddr"`
	VersionHeader    string            `mapstructure:"versionHeader"`
	VersionPorts     map[string]int    `mapstructure:"versionPorts"`
	Routes           []Route           `mapstructure:"routes"`
	AggregatedRoutes []AggregatedRoute `mapstructure:"aggregatedRoutes"`
	ChainedRoutes    []ChainedRoute    `mapstructure:"chainedRoutes"`
}

type Config struct {
	Gateway GatewayConfig `mapstructure:"gateway"`
}

func LoadConfig(configPath string) (*Config, error) {
	viper.AddConfigPath(configPath)
	viper.SetConfigType("yaml")
	viper.SetConfigName("default")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
