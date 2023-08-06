package http

import (
	"net/http"
	"net/url"
)

// transportOptions contains transport specific configuration.
type transportOptions struct {
	insecureSkipTLS bool
	// []byte is not comparable.
	caBundle string
	proxyURL url.URL
}

func (c *client) addTransport(opts transportOptions, transport *http.Transport) {
	c.m.Lock()
	c.transports.Add(opts, transport)
	c.m.Unlock()
}

func (c *client) removeTransport(opts transportOptions) {
	c.m.Lock()
	c.transports.Remove(opts)
	c.m.Unlock()
}

func (c *client) fetchTransport(opts transportOptions) (*http.Transport, bool) {
	c.m.RLock()
	t, ok := c.transports.Get(opts)
	c.m.RUnlock()
	if !ok {
		return nil, false
	}
	transport, ok := t.(*http.Transport)
	if !ok {
		return nil, false
	}
	return transport, true
}
