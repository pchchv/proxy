package main

import (
	"crypto/sha256"
	"hash"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type Cache struct {
	folder      string
	hash        hash.Hash
	knownValues map[string][]byte
	busyValues  map[string]*sync.Mutex
	mutex       *sync.Mutex
}

func CreateCache(path string) (*Cache, error) {
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("Cannot open cache folder '%s': %s", path, err)
		log.Printf("Create cache folder '%s'", path)
		os.Mkdir(path, os.ModePerm)
	}

	values := make(map[string][]byte, 0)
	busy := make(map[string]*sync.Mutex, 0)

	// Go through each file and save its name to the map. The contents of the file are loaded as needed.
	// This keeps us from having to read the contents of the directory every time the user needs data that has not yet been loaded.
	for _, info := range fileInfos {
		if !info.IsDir() {
			values[info.Name()] = nil
		}
	}

	hash := sha256.New()

	mutex := &sync.Mutex{}

	cache := &Cache{
		folder:      path,
		hash:        hash,
		knownValues: values,
		busyValues:  busy,
		mutex:       mutex,
	}

	return cache, nil
}
