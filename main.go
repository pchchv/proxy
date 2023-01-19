package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

var (
	config *Config
	cache  *Cache
	client *http.Client
)

func prepare() {
	var err error

	cache, err = CreateCache(config.CacheFolder)

	if err != nil {
		log.Fatalf("Could not init cache: '%s'", err.Error())
	}

	client = &http.Client{
		Timeout: time.Second * 30,
	}
}

func main() {
	var err error

	configPath := flag.String("config", "./conf.json", "configuration .json file path")
	flag.Parse()

	config, err = LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Could not read config: '%s'", err.Error())
	}

	prepare()
}
