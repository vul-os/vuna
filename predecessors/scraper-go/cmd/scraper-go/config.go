package main

import (
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	err := viper.BindEnv("ServerPort", "PORT")
	if err != nil {
		log.Fatal().Err(err).Msgf("binding environment variables to configuration keys")
	}

	// set default configuration
	viper.SetDefault("ServerPort", "8080")
}

type Config struct {
	ServerPort string

	// main db
	DatabaseName                   string
	DatabaseUser                   string
	DatabasePassword               string
	DatabaseHost                   string
	DatabasePort                   string
	DatabaseInstanceConnectionName string // for gcp cloud storage

	// gcp bigquery
	BQClientProjectID string
}

func GetConfig(configFileName *string) (*Config, error) {
	// set places to look for config file

	// local
	viper.AddConfigPath("cmd" + string(os.PathSeparator) + "scraper-go")
	viper.AddConfigPath(".")

	// cloud run
	viper.AddConfigPath("../../config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("./config")

	// set the name of the config file
	viper.SetConfigName(*configFileName)
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msgf("could not parse config file")
		return nil, err
	}

	// parse the config file
	cfg := new(Config)
	if err := viper.Unmarshal(cfg); err != nil {
		log.Error().Err(err).Msg("unmarshalling config file")
		return nil, err
	}

	return cfg, nil
}
