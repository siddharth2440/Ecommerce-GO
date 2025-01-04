package config

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewDB(cfg *Config) (*mongo.Client, error) {
	mongo_uri := cfg.MONGO_URI
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	optns := options.Client().ApplyURI(mongo_uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), optns)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	mongoClient := client
	fmt.Println("DataBase Connected Successfully!")
	return mongoClient, nil
}
