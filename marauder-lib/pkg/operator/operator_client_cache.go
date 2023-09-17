package operator

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
)

// The ClientCache holds the different Client instances.
type ClientCache struct {
	sharedClient *http.Client
	protocol     string
	clients      map[string]Client
}

// NewOperatorClientCache creates a new operator client cache.
func NewOperatorClientCache(sharedClient *http.Client, protocol string) *ClientCache {
	return &ClientCache{sharedClient: sharedClient, protocol: protocol, clients: make(map[string]Client)}
}

// GetOrCreate constructs a new operator client or returns the cached one.
func (o *ClientCache) GetOrCreate(identifier string, host string, port int) Client {
	client, ok := o.clients[identifier]
	if ok {
		return client
	}

	httpClient := &HTTPClient{
		Client:      o.sharedClient,
		OperatorURL: fmt.Sprintf("%s://%s", o.protocol, net.JoinHostPort(host, strconv.Itoa(port))),
	}
	o.clients[identifier] = httpClient

	return httpClient
}
