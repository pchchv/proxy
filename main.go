package main

import (
	"flag"
	"fmt"
	"io"
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

func handleGet(w http.ResponseWriter, r *http.Request) {
	fullUrl := r.URL.Path + "?" + r.URL.RawQuery

	log.Printf("Requested '%s'", fullUrl)

	// Cache miss -> Load data from requested URL and add to cache
	if busy, ok := cache.has(fullUrl); !ok {
		defer busy.Unlock()
		response, err := client.Get(config.Target + fullUrl)
		if err != nil {
			handleError(err, w)
			return
		}

		var reader io.Reader
		reader = response.Body

		err = cache.put(fullUrl, &reader, response.ContentLength)
		if err != nil {
			handleError(err, w)
			return
		}
		defer response.Body.Close()
	}

	// The cache has definitely the data we want, so get a reader for that
	content, err := cache.get(fullUrl)

	if err != nil {
		handleError(err, w)
	} else {
		_, err := io.Copy(w, *content)
		if err != nil {
			log.Fatalf("Error writing response: %s", err.Error())
			handleError(err, w)
			return
		}
	}
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
