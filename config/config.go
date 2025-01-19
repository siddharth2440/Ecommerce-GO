package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// defining envireoment variables Structure
type Config struct {
	MONGO_URI   string
	PORT        string
	JWT_SECRET  string
	UPSTASH_URI string
}

func SetConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()

	if err != nil {
		return nil, err
	}

	port := viper.GetString("PORT")
	fmt.Printf("PORT from the .env : %s", port)
	fmt.Println("Hii this is It'sMoNdAy")

	db_uri := viper.GetString("DB_URI")
	fmt.Printf("Database URI : %s", db_uri)
	return &Config{
		MONGO_URI:   viper.GetString("DB_URI"),
		PORT:        viper.GetString("PORT"),
		JWT_SECRET:  viper.GetString("JWT_SECRET"),
		UPSTASH_URI: viper.GetString("UPSTASH_URI"),
	}, nil
}
