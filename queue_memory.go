package kvmq

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"time"
)

type MemoryQueue struct {
	Name  string
	index []string
	queue map[string]*QueueObject
}

func newMemoryQueue(config *Config) (*MemoryQueue, error) {
	_ = config
	return &MemoryQueue{}, nil
}

func (mq *MemoryQueue) Connect() error {
	mq.index = []string{}
	mq.queue = map[string]*QueueObject{}
	return nil
}

func (mq *MemoryQueue) Length() (int, error) {
	return len(mq.index), nil
}

func (mq *MemoryQueue) PushIndex(body []byte, index int) (object *QueueObject, err error) {
	if index > len(mq.index)-1 {
		return nil, fmt.Errorf("index %v out of bounds", index)
	}
	qo := &QueueObject{
		ID:        mq.hash(body),
		CreatedAt: time.Now(),
		Body:      body,
	}
	mq.queue[qo.ID] = qo
	switch index {
	case -1:
		mq.index = append(mq.index, qo.ID)
	case 0:
		mq.index = append([]string{qo.ID}, mq.index...)
	default:
		mq.index = slices.Insert(mq.index, index, qo.ID)
	}
	return qo, nil
}

func (mq *MemoryQueue) PopIndex(index int) (body []byte, object *QueueObject, err error) {
	if index > len(mq.index)-1 {
		return []byte{}, nil, fmt.Errorf("index %v out of bounds", index)
	}
	var ID string
	switch index {
	case -1:
		ID = mq.index[:len(mq.index)-1][0]
	default:
		ID = mq.index[index]
		mq.index = append(mq.index[:index], mq.index[index:]...)
	}
	qo := mq.queue[ID]
	delete(mq.queue, ID)
	return qo.Body, qo, nil
}

func (mq *MemoryQueue) PeekIndex(index int) (body []byte, object *QueueObject, err error) {
	if index > len(mq.index)-1 {
		return []byte{}, nil, fmt.Errorf("index %v out of bounds", index)
	}
	var ID string
	switch index {
	case -1:
		ID = mq.index[len(mq.index)-1]
	default:
		ID = mq.index[index]
	}
	qo := mq.queue[ID]
	return qo.Body, qo, nil
}

func (mq *MemoryQueue) PeekScan() (bodies [][]byte, objects []*QueueObject, err error) {
	return [][]byte{}, nil, nil
}

func (mq *MemoryQueue) Find(match []byte) (found bool, index int, object *QueueObject, err error) {
	return false, 0, nil, nil
}

func (mq *MemoryQueue) ClearQueue() error {
	mq.index = []string{}
	mq.queue = map[string]*QueueObject{}
	return nil
}

func (mq *MemoryQueue) RebuildIndex() error {
	mq.index = []string{}
	for k := range mq.queue {
		mq.index = append(mq.index, k)
	}
	return nil
}

func (mq *MemoryQueue) DeleteQueue() error {
	return mq.ClearQueue()
}

func (mq MemoryQueue) hash(item []byte) string {
	h := sha1.New()
	io.WriteString(h, string(item[:]))
	return string(h.Sum(nil))
}

func (mq *MemoryQueue) DebugIndex() {
	b, _ := json.MarshalIndent(mq.index, "", " ")
	fmt.Println(b)
}

func (mq *MemoryQueue) DebugQueue() {
	b, _ := json.MarshalIndent(mq.queue, "", " ")
	fmt.Println(b)
}
