package config

type Config struct {
	DBConnectionString string
	PubSubSubscription string
}

func Load() (*Config, error) {
	// Load config using Viper or another config management tool
	return &Config{
		DBConnectionString: "your_connection_string",
		PubSubSubscription: "scan-sub",
	}, nil
}
