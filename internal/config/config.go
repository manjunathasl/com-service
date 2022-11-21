package config

import "os"

type Config struct {
	CustomerFilePath string
	ComPubService    string
}

//constructor for config
func New() Config {
	cfg := Config{}

	if v := os.Getenv("CUSTOMERS_FILE_PATH"); v != "" {
		cfg.CustomerFilePath = v
	} else {
		cfg.CustomerFilePath = "./customers.csv"
	}

	if v := os.Getenv("COM_PUB_SERVICE"); v != "" {
		cfg.ComPubService = v
	} else {
		cfg.ComPubService = "http://localhost:9090/messages"
	}

	return cfg
}
