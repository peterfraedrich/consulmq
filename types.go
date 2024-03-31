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

	// Removes an item by its ID
	PopID(id string) (body []byte, object *QueueObject, err error)

	// Gets an item at index but leaves it in the queue; index follows same pattern as above
	PeekIndex(index int) (body []byte, object *QueueObject, err error)

	// Gets an item by its ID but leaves it in the queue
	PeekID(id string) (body []byte, object *QueueObject, err error)

	// Gets all items in the queue but leaves them in place; returns a map of index:object
	PeekScan() (bodies [][]byte, objects map[int]*QueueObject, err error)

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
	Name string `default:"kvmq"`
	// Backend to use
	Backend string `default:"memory"`
	// Seconds a lock should be valid for
	LockTTL int `default:"5"`
	// Seconds to wait for a lock to come available
	LockTimeout int `default:"30"`
	// Config values for Memory backend
	MemoryConfig interface{}
	// Config values for Redis backend
	RedisConfig RedisConfig
	// Config values for Consul backend
	RDBMSConfig RDBMSConfig
	// Config values for Filesystem backend
	FilesystemConfig FilesystemConfig
}

type RedisConfig struct {
	DB int `default:"0"`
}

type RDBMSConfig struct {
	// The type of Database Engine to use (Postgres, MySQL, SQLite, etc.)
	Engine string `default:"sqlite"`
	// Path to SQLite DB to use
	SQLiteDB string `default:"kvmq.db"`
	// Connection String
	ConnString string
	// Enable hard delete
	HardDelete bool `default:"false"`
}

type FilesystemConfig struct {
	// base directory for the queue files
	Directory string `default:".kvmq/"`
}

// QueueObject is a container around any data in the queue
type QueueObject struct {
	// Unique ID of the object
	ID any
	// Creation time of the object
	CreatedAt time.Time
	// When the object will be deleted
	TTLDeadline time.Time
	// Any tags for the object (TBI)
	// TODO: Implement message tagging
	// Tags []string
	// The actual data to be put on the queue
	Body []byte
}

type ID interface {
	uint | string
}

type Lock struct {
	PID int
	TTL time.Time
}
