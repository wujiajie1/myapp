package cluster

import (
	"errors"
	"sync/atomic"
	"vendor"
)

var errClientInUse = errors.New("cluster: client is already used by another consumer")

// Client is a group client
type Client struct {
	vendor.Client
	config vendor.Config

	inUse uint32
}

// NewClient creates a new client instance
func NewClient(addrs []string, config *vendor.Config) (*Client, error) {
	if config == nil {
		config = vendor.NewConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	client, err := vendor.NewClient(addrs, &config.Config)
	if err != nil {
		return nil, err
	}

	return &Client{Client: client, config: *config}, nil
}

// ClusterConfig returns the cluster configuration.
func (c *Client) ClusterConfig() *vendor.Config {
	cfg := c.config
	return &cfg
}

func (c *Client) claim() bool {
	return atomic.CompareAndSwapUint32(&c.inUse, 0, 1)
}

func (c *Client) release() {
	atomic.CompareAndSwapUint32(&c.inUse, 1, 0)
}
