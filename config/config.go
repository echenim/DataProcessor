package config

type Config struct {
	DBConnectionString string
	PubSubSubscription string
}

func Load(conn_string, scan_topic string) (*Config, error) {
	// Load config using Viper or another config management tool
	return &Config{
		DBConnectionString: conn_string,
		PubSubSubscription: scan_topic,
	}, nil
}
