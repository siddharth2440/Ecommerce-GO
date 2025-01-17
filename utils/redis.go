package utils

import (
	"github.com/redis/go-redis"
)

func Get_Redis() *redis.Client {
	// Take your own Redis URI then then paste here i.e., inside the PasrseURL
	opt, _ := redis.ParseURL("rediss://default:********@inspired-mudfish-27405.upstash.io:6379")
	client := redis.NewClient(opt)
	return client
}
