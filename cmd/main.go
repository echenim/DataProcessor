package main

import (
	"context"
	"log"

	//	"time"

	"github.com/echenim/data-processor/services"

	"github.com/echenim/data-processor/repositories"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/viper"
)

func initConfig() {
	// Sets the name of the configuration file without its file extension.
	// This tells Viper what file to look for. In this case, "config/config" implies
	// Viper will look for a file named "config.yaml" or "config.json" depending on the set type in the specified paths.
	viper.SetConfigName("config/config")

	// Specifies the type of the configuration file. This example sets it to YAML,
	// but it can be changed to JSON or other supported types by Viper by adjusting this line.
	viper.SetConfigType("yaml")

	// Adds the current directory (".") to the list of paths Viper will search for the configuration file.
	// You can add multiple paths here if your configuration file might be in different locations.
	viper.AddConfigPath(".")

	// Reads in the configuration file using the previously set name, type, and path(s).
	// This method looks for the configuration file in the specified path and parses it.
	err := viper.ReadInConfig()
	// If Viper encounters an error while reading the configuration file (e.g., file not found, parse errors),
	// it will terminate the application with a fatal error, logging the encountered issue.
	if err != nil {
		log.Fatalf("Fatal error config file: %s \n", err)
	}
}

func main() {
	// Initializes the application configuration by reading from the config file.
	// This setup must be done before accessing any configuration values.
	initConfig()

	// Initializes a new PostgreSQL client with the database driver and data source name
	// specified in the configuration file. This is typically a connection string.
	// Errors during client creation are logged, not causing the application to exit.
	db, err := repositories.PostgreSQlProviderClient(viper.GetString("database.driver"), viper.GetString("database.dataSourceName"))
	if err != nil {
		log.Printf("\n Error : %v", err)
	}

	// Initializes a new Pub/Sub client with the project ID specified in the configuration file.
	// This client will be used for publishing or subscribing to messages within the application.
	// Similar to the database client, errors here are logged.
	pubsub, err := repositories.PubSubProviderClient(viper.GetString("pubsub.project_id"))
	if err != nil {
		log.Printf("\n Error : %v", err)
	}

	// Creates a new instance of ScannedProcessorRepository using the previously
	// initialized Pub/Sub and database clients. This repository abstracts away
	// the data access and messaging logic for scanned data.
	repo := repositories.NewScannedProcessorRepository(pubsub, db)

	// Initializes the ScannedProcessorService with the repository.
	// The service is responsible for the business logic related to processing scanned data.
	srv := services.NewScannedProcessorService(repo)

	// Calls the service method to process scanned data. This method will likely
	// involve reading data from a source, processing it, and performing database
	// operations or publishing messages. The context passed here is the background context,
	// indicating that this operation is not meant to be canceled and is essential
	// for the application's core functionality.
	ctx := context.Background()
	subscriptionID := viper.GetString("pubsub.scan_subscription")
	batchSize := viper.GetInt("concurrency.batch_Size")
	batchTimeout := viper.GetDuration("concurrency.batch_Timeout")

	srv.ProcessScanData(ctx, subscriptionID, batchSize, batchTimeout)
}
