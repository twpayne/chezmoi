package cmd

import (
	"net/url"
	"sync"

	"github.com/bartventer/httpcache/store"
	"github.com/bartventer/httpcache/store/driver"
	"github.com/bartventer/httpcache/store/fscache"
)

const httpCacheScheme = "httpcache"

// httpCache implements [github.com/bartventer/httpcache/store/driver.Conn]. It
// exists to give precise control over the cache directory location, which is
// not available with fscache.
type httpCache struct {
	open func() (driver.Conn, error)
}

// Get implements [github.com/bartventer/httpcache/store/driver.Conn.Get].
func (c *httpCache) Get(key string) ([]byte, error) {
	conn, err := c.open()
	if err != nil {
		return nil, err
	}
	return conn.Get(key)
}

// Set implements [github.com/bartventer/httpcache/store/driver.Conn.Set].
func (c *httpCache) Set(key string, entry []byte) error {
	conn, err := c.open()
	if err != nil {
		return err
	}
	return conn.Set(key, entry)
}

// Delete implements [github.com/bartventer/httpcache/store/driver.Conn.Delete].
func (c *httpCache) Delete(key string) error {
	conn, err := c.open()
	if err != nil {
		return err
	}
	return conn.Delete(key)
}

func init() {
	store.Register(httpCacheScheme, driver.DriverFunc(func(u *url.URL) (driver.Conn, error) {
		open := sync.OnceValues(func() (driver.Conn, error) {
			return fscache.Open(httpCacheScheme,
				fscache.WithBaseDir(u.Path),
				fscache.WithUpdateMTime(true),
			)
		})
		return &httpCache{open: open}, nil
	}))
}
