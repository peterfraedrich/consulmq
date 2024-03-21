package kvmq

import (
	"time"
)

type Backend interface {
	// Sets up any connections, filesystems, etc. that the queue needs.
	// this is called every time a new Queue is created and is the first method called.
	Connect() error

	// Returns the number of items in the queue
	Length() (int, error)

	// Push an item to a specific index. This is called by Push, PushFirst, and PushIndex.
	// index = 0 -- push item to the front of the queue
	// index = n -- push item to index n
	// index = -1 -- push item to end of queue
	PushIndex(body []byte, index int) (object *QueueObject, err error)

	// Removes an item from the queue; index follows same pattern as above
	PopIndex(index int) (body []byte, object *QueueObject, err error)

	// Gets an item at index but leaves it in the queue; index follows same pattern as above
	PeekIndex(index int) (body []byte, object *QueueObject, err error)

	// Gets all items in the queue but leaves them in place
	PeekScan() (bodies [][]byte, objects []*QueueObject, err error)

	// Searches for the first item with an exact match
	Find(match []byte) (found bool, index int, object *QueueObject, err error)

	// Removes all items from a queue but leaves the queue iteself intact
	ClearQueue() error

	// Rebuilds the index from the items in the queue
	RebuildIndex() error

	// Deletes the queue and all items associated with it
	DeleteQueue() error

	DebugIndex()
	DebugQueue()
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
