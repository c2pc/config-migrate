package main

import (
	"fmt"
	"log"
)

const configPath = "config.yaml"

func main() {
	//Run config migrations
	if err := runMigrationWithComments(configPath); err != nil {
		log.Fatalf("Error running migration: %s", err)
	}

	//Parse config file after migrate
	cfg, err := parseConfig(configPath)
	if err != nil {
		log.Fatalf("Error parsing config: %s", err)
	}

	//Print results
	fmt.Printf("%+v\n", cfg)
}
