package kvmq

import (
	"time"
)

type Backend interface {
	Connect() error
}

// Config is for passing configuration into the Connect function
type Config struct {
	// Unique name of the message queue
	Name string
	// Backend to use
	Backend string
	// Config values for Memory backend
	MemoryConfig struct{}
	// Config values for Redis backend
	RedisConfig struct {
		DB int
	}
	// Config values for Consul backend
	ConsulConfig struct {
		Datacenter string
	}
	// Config values for Filesystem backend
	FilesystemConfig struct {
		// base directory for the queue files
		Directory string
	}
}

type queue struct {
	Name       string
	RootPath   string        `json:"root_path"`
	SystemPath string        `json:"system_path"`
	QueuePath  string        `json:"queue_path"`
	RetryPath  string        `json:"retry_path"`
	CreatedAt  time.Time     `json:"created_at"`
	TTL        time.Duration `json:"ttl"`
}

var defaults = map[string]string{
	"Address":    "localhost:8500",
	"Datacenter": "dc1",
	"MQName":     "consulmq",
}

// QueueObject is a container around any data in the queue
type QueueObject struct {
	// Unique ID of the object
	ID string
	// Creation time of the object
	CreatedAt time.Time
	// When the object will be deleted
	TTLDeadline time.Time
	// Any tags for the object (TBI)
	// TODO: Implement message tagging
	Tags []string
	// The actual data to be put on the queue
	Body []byte
}
