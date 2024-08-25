package config

import (
	"github.com/spf13/viper"
)

// Env to manage config
type Env struct {
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
	TcpPort       string `mapstructure:"TCP_PORT"`
	LogEncoding   string
}

// AppConfig holding env
var AppConfig Env

// InitConfig which inits config
func InitConfig() error {

	//  current path
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	//using viper for reading env
	err := viper.Unmarshal(&AppConfig)

	if err != nil {
		return err
	}
	return nil
}
