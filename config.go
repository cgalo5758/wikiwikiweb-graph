package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func configure() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No config file found. Provide a config.yaml file.")
		} else {
			fmt.Println("Error reading config file: ", err.Error())
		}
	}
}
