package config_loader

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	db "web-server/db"
	ping_dto "web-server/db/dto"

	"github.com/spf13/viper"
)

type TargetConfig struct {
	Addr            string `mapstructure:"addr"`
	Context         string `mapstructure:"context"`
	Port            int    `mapstructure:"port,omitempty"`
	ProxyPortHeader string `mapstructure:"proxyPortHeader,omitempty"`
}

func (t *TargetConfig) GetURL(r *http.Request) (string, error) {
	if t.Port > 0 {
		return fmt.Sprintf("%s:%d%s", t.Addr, t.Port, t.Context), nil
	}

	if t.ProxyPortHeader != "" {
		headerValue := r.Header.Get(t.ProxyPortHeader)
		if headerValue == "" {
			return "", fmt.Errorf("header '%s' is not present in the request", t.ProxyPortHeader)
		}

		portDto, err := ping_dto.GetPortByHeaderAndValue(db.GetDB(), t.ProxyPortHeader, headerValue)
		if err != nil {
			return "", fmt.Errorf("failed to get port from database: %v", err)
		}

		if portDto == nil {
			return "", fmt.Errorf("no port found in the database for header '%s' with value '%s'", t.ProxyPortHeader, headerValue)
		}

		return fmt.Sprintf("%s:%d%s", t.Addr, portDto.Port, t.Context), nil
	}

	return "", fmt.Errorf(" neither 'port' nor 'proxyPortHeader' is set")
}

func (t *TargetConfig) Validate() error {
	if (t.Port != 0 && t.ProxyPortHeader != "") || (t.Port == 0 && t.ProxyPortHeader == "") {
		return errors.New("either 'port' or 'proxyPortHeader' must be set, but not both")
	}
	return nil
}

type Route struct {
	Name    string       `mapstructure:"name"`
	Context string       `mapstructure:"context"`
	Target  TargetConfig `mapstructure:"target"`
}

type AggregatedRoute struct {
	Name    string         `mapstructure:"name"`
	Context string         `mapstructure:"context"`
	Targets []TargetConfig `mapstructure:"targets"`
}

type ChainedRoute struct {
	Name    string         `mapstructure:"name"`
	Context string         `mapstructure:"context"`
	Targets []TargetConfig `mapstructure:"targets"`
}

type GroupeRoute struct {
	Name          string       `mapstructure:"name"`
	ContextPrefix string       `mapstructure:"contextPrifix"`
	Target        TargetConfig `mapstructure:"target"`
}

type GatewayConfig struct {
	ListenAddr       string            `mapstructure:"listenAddr"`
	Routes           []Route           `mapstructure:"routes"`
	AggregatedRoutes []AggregatedRoute `mapstructure:"aggregatedRoutes"`
	ChainedRoutes    []ChainedRoute    `mapstructure:"chainedRoutes"`
	GroupeRoutes     []GroupeRoute     `mapstructure:"groupRoutes"`
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

	log.Printf("Validate route...")
	for _, route := range config.Gateway.Routes {
		if err := route.Target.Validate(); err != nil {
			return nil, err
		}
	}

	log.Printf("Validate agregated...")
	for _, aggregatedRoute := range config.Gateway.AggregatedRoutes {
		for _, target := range aggregatedRoute.Targets {
			if err := target.Validate(); err != nil {
				return nil, err
			}
		}
	}

	log.Printf("Validate chained...")
	for _, chainedRoute := range config.Gateway.ChainedRoutes {
		for _, target := range chainedRoute.Targets {
			if err := target.Validate(); err != nil {
				return nil, err
			}
		}
	}

	log.Printf("Validate group...")
	for _, webSocketRoute := range config.Gateway.GroupeRoutes {
		if err := webSocketRoute.Target.Validate(); err != nil {
			return nil, err
		}
	}

	return &config, nil
}
