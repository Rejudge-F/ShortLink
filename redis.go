package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/catinello/base62"
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
	URL                 string `json:"url"`
	CreatedAt           string `json:"created_at"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes"`
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

func (cli *RedisClient) Shorten(url string, exp int64) (string, error) {
	hashUrl := fmt.Sprintf("%s", sha1.Sum([]byte(url)))

	d, err := cli.Cli.Get(hashUrl).Result()

	if err == redis.Nil {
		// url not exist, nothing to do
	} else if err != nil {
		return "", err
	} else {
		if d == "{}" {
			// url expired, nothing to do
		} else {
			return d, nil
		}
	}

	// Incr global counter
	if err := cli.Cli.Incr(URLIDKEY).Err(); err != nil {
		return "", err
	}

	// get global counter
	urlId, err := cli.Cli.Get(URLIDKEY).Int()
	if err != nil {
		return "", err
	}

	// convert int to short link
	eid := base62.Encode(urlId)

	// set key fot short link to url
	err = cli.Cli.Set(fmt.Sprintf(ShortlinkKey, eid), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	// set key for sha1(url) to short link
	err = cli.Cli.Set(fmt.Sprintf(URLHashKey, hashUrl), eid, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	// set detail for short link
	shortLinkInfo := &ShortlinkInfo{
		URL:                 url,
		ExpirationInMinutes: exp,
		CreatedAt:           time.Now().String(),
	}

	// serialize short link info
	jsonStr, err := json.Marshal(shortLinkInfo)
	if err != nil {
		return "", err
	}

	// set key for short link to detail
	err = cli.Cli.Set(fmt.Sprintf(ShortlinkDetailKey, eid), jsonStr, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	return eid, err
}

func (cli *RedisClient) ShortlinkInfo(eid string) (interface{}, error) {
	return nil, nil
}

func (cli *RedisClient) UnShorten(eid string) (string, error) {
	return "", nil
}
