package config

import (
	"github.com/spf13/viper"
)

type GoogleConfig struct {
	ClientID     string `mapstructure:"client_id"`
	Secret       string `mapstructure:"secret"`
	DiscoveryURL string `mapstructure:"discovery_url"`
	Callback     string `mapstructure:"callback"`
}

type Config struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	JWTSecret string `mapstructure:"jwt_secret"`
	Google GoogleConfig `mapstructure:"google"`
}

func New(configFile string) (*Config, error) {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))

	if configFile == "" {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
	} else {
		v.SetConfigFile(configFile)
	}

	v.SetEnvPrefix("tmp")
	v.AutomaticEnv()

	var config = Config{}

	if err := v.ReadInConfig(); err != nil {
		return &config, nil
	}

	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
