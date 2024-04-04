package main

import (
	"context"
	"log"

	"github.com/echenim/data-processor/clients"
	"github.com/echenim/data-processor/config"
	"github.com/echenim/data-processor/logger"
	"github.com/echenim/data-processor/processor"
	"github.com/spf13/viper"
)

func main() {
	initConfig()

	logger.Setup(viper.GetString("log.root_directory"))
	cfg, err := config.Load(viper.GetString("database.dataSourceName"), viper.GetString("pubsub.scan_topic"))
	if err != nil {
		logger.Error("Failed to load configuration:", err)
		return
	}

	ctx := context.Background()
	processor, err := processor.NewProcessor(ctx, cfg)
	if err != nil {
		logger.Error("Failed to initialize processor:", err)
		return
	}

	clients.StartSubscriber(ctx, cfg, processor.ProcessMessage)
}

func initConfig() {
	viper.SetConfigName("config/config") // name of config file (without extension)
	viper.SetConfigType("yaml")          // or viper.SetConfigType("json")
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	err := viper.ReadInConfig()          // Find and read the config file
	if err != nil {                      // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s \n", err)
	}
}
