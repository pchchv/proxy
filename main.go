package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	config *Config
	cache  *Cache
	client *http.Client
)

func loadConfig(configPath string) error {
	var err error

	config, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("Could not read config: %s", err)
	}

	return nil
}

func prepare() error {
	var err error

	cache, err = CreateCache(config.CacheFolder)
	if err != nil {
		return fmt.Errorf("Could not init cache: '%s'", err)
	}

	client = &http.Client{
		Timeout: time.Second * 30,
	}

	return nil
}

func handleError(err error, w http.ResponseWriter) {
	log.Fatal(err.Error())
	w.WriteHeader(500)
	fmt.Fprintf(w, err.Error())
}

func main() {
	configPath := flag.String("config", "./conf.json", "configuration .json file path")
	flag.Parse()

	err := loadConfig(*configPath)
	if err != nil {
		log.Panic(err)
	}

	err = prepare()
	if err != nil {
		log.Panic(err)
	}
}
