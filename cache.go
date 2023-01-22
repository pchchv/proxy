package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/pchchv/golog"
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
		golog.Error("Cannot open cache folder '%s': %s", path, err)
		golog.Info("Create cache folder '%s'", path)
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

func (c *Cache) get(key string) (*io.Reader, error) {
	var response io.Reader
	hashValue := calcHash(key)

	// Try to get content. Error if not found.
	c.mutex.Lock()
	content, ok := c.knownValues[hashValue]
	c.mutex.Unlock()
	if !ok && len(content) > 0 {
		golog.Debug("Cache doesn't know key '%s'", hashValue)
		return nil, fmt.Errorf("Key '%s' is not known to cache", hashValue)
	}

	// Key is known, but not loaded into RAM
	if content == nil {
		golog.Debug("Cache item '%s' known but is not stored in memory. Using file.", hashValue)

		file, err := os.Open(c.folder + hashValue)
		if err != nil {
			golog.Error("Error reading cached file '%s': %s", hashValue, err)
			return nil, err
		}

		response = file

		golog.Debug("Create reader from file %s", hashValue)
	} else {
		// Key is known and data is already loaded to RAM
		response = bytes.NewReader(content)

		golog.Debug("Create reader from %d byte large cache content", len(content))
	}

	return &response, nil
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

func (c *Cache) put(key string, content *io.Reader, contentLength int64) error {
	hashValue := calcHash(key)

	// Small enough to put it into the in-memory cache
	if contentLength <= config.MaxCacheItemSize*1024*1024 {
		buffer := &bytes.Buffer{}
		_, err := io.Copy(buffer, *content)
		if err != nil {
			return err
		}

		defer c.release(hashValue, buffer.Bytes())
		golog.Debug("Added %s into in-memory cache", hashValue)

		if err = ioutil.WriteFile(c.folder+hashValue, buffer.Bytes(), 0o644); err != nil {
			return err
		}

		golog.Debug("Wrote content of entry %s into file", hashValue)
	} else {
		// Too large for in-memory cache, just write to file
		defer c.release(hashValue, nil)
		golog.Debug("Added nil-entry for %s into in-memory cache", hashValue)

		file, err := os.Create(c.folder + hashValue)
		if err != nil {
			return err
		}

		writer := bufio.NewWriter(file)
		_, err = io.Copy(writer, *content)
		if err != nil {
			return err
		}

		golog.Debug("Wrote content of entry %s into file", hashValue)
	}

	golog.Debug("Cache wrote content into '%s'", hashValue)

	return nil
}

// Internal method which atomically caches an item and unmarks the item as busy,
// if it was busy before. The busy lock *must* be unlocked elsewhere!
func (c *Cache) release(hashValue string, content []byte) {
	c.mutex.Lock()
	delete(c.busyValues, hashValue)
	c.knownValues[hashValue] = content
	c.mutex.Unlock()
}

func calcHash(data string) string {
	sha := sha256.Sum256([]byte(data))

	return hex.EncodeToString(sha[:])
}
