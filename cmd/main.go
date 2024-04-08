package main

import (
	"context"
	"log"

	"github.com/echenim/data-processor/services"

	"github.com/echenim/data-processor/repositories"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetConfigName("config/config") // name of config file (without extension)
	viper.SetConfigType("yaml")          // or viper.SetConfigType("json")
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	err := viper.ReadInConfig()          // Find and read the config file
	if err != nil {                      // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s \n", err)
	}
}

func main() {
	initConfig()

	db, err := repositories.PostgreSQlProviderClient(viper.GetString("database.driver"), viper.GetString("database.dataSourceName"))
	if err != nil {
		log.Printf("\n Error : %v", err)
	}

	pubsub, err := repositories.PubSubProviderClient(viper.GetString("pubsub.project_id"))
	if err != nil {
		log.Printf("\n Error : %v", err)
	}

	repo := repositories.NewScannedProcessorRepository(pubsub, db)

	srv := services.NewScannedProcessorService(repo)

	srv.ProcessScanData(context.Background())
}
