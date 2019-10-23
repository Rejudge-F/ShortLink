package main

import (
	"github.com/astaxie/beego/logs"
	"os"
	"strconv"
)

type Env struct {
	storage Storage
}

func getEnv() *Env {
	addr := os.Getenv("APP_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	pass := os.Getenv("APP_REDIS_PASSWD")
	if pass == "" {
		pass = ""
	}

	db := os.Getenv("APP_REDIS_DB")
	if db == "" {
		db = "0"
	}
	index, err := strconv.Atoi(db)
	if err != nil {
		logs.Debug("%s not db index", db)
		panic(err)
	}

	cli := NewRedisCli(addr, pass, index)

	return &Env{storage: cli}
}
