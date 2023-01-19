package main

import (
	"flag"
	"log"
)

var config *Config

func main() {
	var err error

	configPath := flag.String("config", "./conf.json", "configuration .json file path")
	flag.Parse()

	config, err = LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Could not read config: '%s'", err.Error())
	}
}
