package config

import (
	"log"

	"github.com/spf13/viper"
)

var AppConfig Config

type Config struct {
	Name        string
	Port        int
	Environment string
	Debug       bool
	Version     string
	ApiKey      string
	IP          string
	AdminEmail  string
}

func LoadAppConfig() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/")
	viper.AllowEmptyEnv(true)
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	/*test*/

	AppConfig.Name = viper.GetString("APP_NAME")
	AppConfig.Version = viper.GetString("APP_VERSION")
	AppConfig.Port = viper.GetInt("PORT")
	AppConfig.Environment = viper.GetString("ENVIRONMENT")
	AppConfig.Debug = viper.GetBool("DEBUG")
	AppConfig.ApiKey = viper.GetString("API_KEY")
	AppConfig.IP = viper.GetString("IP")
	AppConfig.AdminEmail = viper.GetString("ADMIN_EMAIL")

	log.Println("[INIT] configuration loaded")
}
