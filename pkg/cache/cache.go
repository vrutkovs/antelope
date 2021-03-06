package cache

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/vrutkovs/antelope/pkg/gcs"
)

type Cache struct {
	sync.Mutex
	cache map[string][]byte
	ctx   context.Context

	Bucket *storage.BucketHandle
}

func (c *Cache) Get(url string) (io.Reader, error) {
	fmt.Printf("cache: fetching url %s\n", url)
	c.Lock()
	defer c.Unlock()

	// Initialise the map if this is the first call
	if c.cache == nil {
		c.cache = make(map[string][]byte)
		c.ctx = context.Background()
	}

	// Return the cached content if it's available
	if r, ok := c.cache[url]; ok {
		return bytes.NewReader(r), nil
	}

	// Fetch real file from gcs
	b, err := gcs.FetchFile(c.Bucket, c.ctx, url)
	if err != nil {
		return nil, err
	}

	c.cache[url] = b
	fmt.Printf("Saved output in cache %s\n", url)

	return bytes.NewReader(b), nil
}
