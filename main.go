package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pchchv/golog"
)

var (
	config *Config
	cache  *Cache
	client *http.Client
)

func loadConfig(configPath string) (err error) {
	config, err = LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("Could not read config: %s", err.Error())
	}

	return nil
}

func prepare() (err error) {
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

	golog.Info("Requested '%s'", fullUrl)

	// Cache miss -> Load data from requested URL and add to cache
	if busy, ok := cache.has(fullUrl); !ok {
		defer busy.Unlock()
		response, err := client.Get(config.Target + fullUrl)
		if err != nil {
			handleError(err, w)
			return
		}

		var reader io.Reader = response.Body

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
			golog.Error("Error writing response: %s", err.Error())
			handleError(err, w)
			return
		}
	}
}

func handleError(err error, w http.ResponseWriter) {
	golog.Fatal(err.Error())
	w.WriteHeader(500)
	fmt.Fprintf(w, "%v", err.Error())
}

func main() {
	configPath := flag.String("config", "./config.json", "configuration .json file path")
	flag.Parse()

	err := loadConfig(*configPath)
	if err != nil {
		golog.Fatal(err.Error())
	}
	if config.DebugLogging {
		golog.LogLevel = golog.LOG_DEBUG
	}
	golog.Debug("Config loaded")

	err = prepare()
	if err != nil {
		golog.Fatal(err.Error())
	}
	golog.Debug("Cache initialized")

	server := &http.Server{
		Addr:         ":" + config.Port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
		Handler:      http.HandlerFunc(handleGet),
	}

	golog.Info("Start serving...")
	err = server.ListenAndServe()
	if err != nil {
		golog.Fatal(err.Error())
	}
}
