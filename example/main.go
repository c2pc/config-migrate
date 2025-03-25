package main

import (
	"fmt"
	"log"
	"os"
)

const configPath = "config.yaml"

func main() {
	//Run config migrations
	if err := runMigration(configPath); err != nil {
		log.Fatalf("Error running migration: %s", err)
	}

	//Parse config file after migrate
	cfg, err := parseConfig(configPath)
	if err != nil {
		log.Fatalf("Error parsing config: %s", err)
	}

	//Print results
	fmt.Printf("%+v\n", cfg)

	//Delete config file after migration (do not use in production)
	err = os.Remove(configPath)
	if err != nil {
		log.Fatalf("Error removing config file: %s", err)
	}
}
