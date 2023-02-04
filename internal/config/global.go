package config

import (
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	RabbitMQ    RabbitMQConfig `yaml:"rabbitmq"`
	MongoDB     MongoDBConfig  `yaml:"mongodb"`
	Development bool           `yaml:"debug"`

	Port uint16 `yaml:"port"`
}

type RabbitMQConfig struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type MongoDBConfig struct {
	URI string `yaml:"uri"`
}

func LoadGlobalConfig() (config *Config, err error) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	if err = viper.Unmarshal(&config); err != nil {
		return
	}

	return
}
