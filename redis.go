package main

import (
	"github.com/go-redis/redis"
	"time"
)

const (
	// URLIDKEY Redis auto incr key
	URLIDKEY = "next.url.id"

	// ShortlinkKey The Key for short_link to url
	ShortlinkKey = "shortlink:%s:url"

	// URLHashKey The Key for url to short_link
	URLHashKey = "urlhash:%s:url"

	// ShortlinkDetailKey The Key for short_link to short_link detail
	ShortlinkDetailKey = "shortlink:%s:detail"
)

type RedisClient struct {
	Cli *redis.Client
}

// ShortlinkInfo short_link info
type ShortlinkInfo struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

func NewRedisCli(addr string, pass string, db int) *RedisClient {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})
	if _, err := c.Ping().Result(); err != nil {
		panic("create redis client failed")
	}
	return &RedisClient{Cli: c}
}
