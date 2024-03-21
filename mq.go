package kvmq

import "strings"

// MQ provides methods for manipulating the message queue
type MQ struct {
	kv Backend
}

func NewMQ(config *Config) (*MQ, error) {
	mq := &MQ{}
	switch strings.ToLower(config.Backend) {
	case "memory":
		memoryqueue, err := newMemoryQueue(config)
		if err != nil {
			return mq, err
		}
		mq.kv = memoryqueue
	case "consul":
		consul, err := NewConsulQueue(config)
		if err != nil {
			return mq, err
		}
		mq.kv = consul
	case "redis":
		redis, err := NewRedisQueue(config)
		if err != nil {
			return mq, err
		}
		mq.kv = redis
	case "filesystem":
		fs, err := NewFilesystemQueue(config)
		if err != nil {
			return mq, err
		}
		mq.kv = fs
	}
	err := mq.kv.Connect()
	if err != nil {
		return mq, err
	}
	return mq, nil
}

func (mq *MQ) Length() (length int, err error) {
	return mq.kv.Length()
}

func (mq *MQ) Push(body []byte) (object *QueueObject, err error) {
	return mq.kv.PushIndex(body, -1)
}

func (mq *MQ) PushFirst(body []byte) (object *QueueObject, err error) {
	return mq.kv.PushIndex(body, 0)
}

func (mq *MQ) PushIndex(body []byte, index int) (object *QueueObject, err error) {
	return mq.kv.PushIndex(body, index)
}

func (mq *MQ) Pop() (body []byte, object *QueueObject, err error) {
	return mq.kv.PopIndex(0)
}

func (mq *MQ) PopLast() (body []byte, object *QueueObject, err error) {
	return mq.kv.PeekIndex(-1)
}

func (mq *MQ) PopIndex(index int) (body []byte, object *QueueObject, err error) {
	return mq.kv.PopIndex(index)
}

func (mq *MQ) Peek() (body []byte, object *QueueObject, err error) {
	return mq.kv.PeekIndex(0)
}

func (mq *MQ) PeekLast() (body []byte, object *QueueObject, err error) {
	return mq.kv.PeekIndex(-1)
}

func (mq *MQ) PeekIndex(index int) (body []byte, object *QueueObject, err error) {
	return mq.kv.PeekIndex(index)
}

func (mq *MQ) PeekScan() (bodies [][]byte, objects []*QueueObject, err error) {
	return mq.kv.PeekScan()
}

func (mq *MQ) Find(match []byte) (found bool, index int, object *QueueObject, err error) {
	return mq.kv.Find(match)
}

func (mq *MQ) ClearQueue() error {
	return mq.kv.ClearQueue()
}

func (mq *MQ) RebuildIndex() error {
	return mq.kv.RebuildIndex()
}

func (mq *MQ) DeleteQueue(confirm bool) error {
	return mq.kv.DeleteQueue()
}

func (mq *MQ) Debug() {
	mq.kv.DebugIndex()
	mq.kv.DebugQueue()
}
