package main

import (
	"crypto/sha256"
	"encoding/hex"
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

// Returns true if the resource is found, and false otherwise.
// If the resource is busy, this method will hang until the resource is free.
// If the resource is not found, a lock indicating that the resource is busy will be returned.
// Once the resource has been put into cache the busy lock *must* be unlocked to allow others to access the newly cached resource.
func (c *Cache) has(key string) (*sync.Mutex, bool) {
	hashValue := calcHash(key)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If the resource is busy, wait for it to be free.
	// This is the case if the resource is currently being cached as a result of another request.
	// Also, release the lock on the cache to allow other readers while waiting.
	if lock, busy := c.busyValues[hashValue]; busy {
		c.mutex.Unlock()
		lock.Lock()
		lock.Unlock()
		c.mutex.Lock()
	}

	// If a resource is in the shared cache, it can't be reserved.
	// One can simply access it directly from the cache.
	if _, found := c.knownValues[hashValue]; found {
		return nil, true
	}

	// The resource is not in the cache, mark the resource as busy until it has been cached successfully.
	// Unlocking lock is required!
	lock := new(sync.Mutex)
	lock.Lock()
	c.busyValues[hashValue] = lock
	return lock, false
}

func calcHash(data string) string {
	sha := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sha[:])
}
