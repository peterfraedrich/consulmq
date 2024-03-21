package kvmq

import "strings"

// MQ provides methods for manipulating the message queue
type MQ struct {
	kv Backend
}

func NewMQ(config Config) (*MQ, error) {
	mq := &MQ{}
	switch strings.ToLower(config.Backend) {
	case "memory":
		memoryqueue, err := newMemoryQueue(config)
		if err != nil {
			return mq, err
		}
		mq.kv = &memoryqueue
	case "consul":
		consul, err := NewConsulQueue(config)
		if err != nil {
			return mq, err
		}
		mq.kv = &consul
	case "redis":
		redis, err := NewRedisQueue(config)
		if err != nil {
			return mq, err
		}
		mq.kv = &redis
	case "filesystem":
		fs, err := NewFilesystemQueue(config)
		if err != nil {
			return mq, err
		}
		mq.kv = &fs
	}
	return mq, nil
}
