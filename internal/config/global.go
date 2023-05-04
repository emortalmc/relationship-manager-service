package config

import (
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	Kafka   *KafkaConfig
	MongoDB *MongoDBConfig

	Development bool
	Port        uint16
}

type KafkaConfig struct {
	Host string
	Port int
}

type MongoDBConfig struct {
	URI string
}

func LoadGlobalConfig() (cfg *Config, err error) {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetConfigName("cfg")
	viper.AddConfigPath(".")

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	if err = viper.Unmarshal(&cfg); err != nil {
		return
	}

	return
}
