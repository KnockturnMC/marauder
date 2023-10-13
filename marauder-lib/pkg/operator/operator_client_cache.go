package operator

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"gitea.knockturnmc.com/marauder/lib/pkg/models/networkmodel"
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

// GetOrCreateFromRef constructs a new operator client or returns the cached one.
func (d *ClientCache) GetOrCreateFromRef(operatorRef networkmodel.ServerOperator) Client {
	return d.GetOrCreate(operatorRef.Identifier, operatorRef.Host, operatorRef.Port)
}

// GetOrCreate constructs a new operator client or returns the cached one.
func (d *ClientCache) GetOrCreate(identifier string, host string, port int) Client {
	client, ok := d.clients[identifier]
	if ok {
		return client
	}

	httpClient := &HTTPClient{
		Client:      d.sharedClient,
		OperatorURL: fmt.Sprintf("%s://%s", d.protocol, net.JoinHostPort(host, strconv.Itoa(port))),
	}
	d.clients[identifier] = httpClient

	return httpClient
}
