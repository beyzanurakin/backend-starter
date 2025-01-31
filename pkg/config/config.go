package config

import (
    "github.com/spf13/viper"
    "log"
)

func InitConfig() {
    viper.SetConfigFile(".env")  // Use an `.env` file
    viper.AutomaticEnv()          // Automatically read env variables

    if err := viper.ReadInConfig(); err != nil {
        log.Printf("Error reading config file: %v", err)
    }
}

func GetEnv(key string) string {
    return viper.GetString(key)
}
