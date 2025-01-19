package utils

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/golang/ecommerce/config"
)

func Get_Redis() *redis.Client {

	// Get the Redis URI
	cfg, _ := config.SetConfig()
	REDIS_URI := cfg.UPSTASH_URI
	fmt.Println(REDIS_URI)

	// Take your own Redis URI then then paste here i.e., inside the PasrseURL
	opt, _ := redis.ParseURL(REDIS_URI)
	client := redis.NewClient(opt)
	return client
}
