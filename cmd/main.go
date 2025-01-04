package main

import (
	"log"

	"github.com/golang/ecommerce/config"
	"github.com/golang/ecommerce/routes"
)

func main() {

	// setup config
	cfg, err := config.SetConfig()
	if err != nil {
		log.Fatalf("Error in Setting up the Configuration file: %v", err)
	}
	// db Connection
	client, err := config.NewDB(cfg)
	if err != nil {
		log.Fatalf("Error in Setting up the DB connection: %v", err)
	}
	// router
	router := routes.SetupRoutes(client)
	router.Run(":8567")
}
