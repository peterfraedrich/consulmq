package kvmq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"time"
)

type MemoryQueue struct {
	Name   string
	qindex []string
	queue  map[string]*QueueObject
	lock   sync.Mutex
}

func newMemoryQueue(config *Config) (*MemoryQueue, error) {
	_ = config
	return &MemoryQueue{}, nil
}

func (mq *MemoryQueue) Connect() error {
	mq.qindex = []string{}
	mq.queue = map[string]*QueueObject{}
	return nil
}

func (mq *MemoryQueue) Length() (int, error) {
	return len(mq.qindex), nil
}

func (mq *MemoryQueue) PushIndex(body []byte, index int) (object *QueueObject, err error) {
	mq.lock.Lock()
	defer mq.deferFunc()
	if index > len(mq.qindex)-1 {
		return nil, fmt.Errorf("index %v out of bounds", index)
	}
	qo := &QueueObject{
		ID:        hash(body),
		CreatedAt: time.Now(),
		Body:      body,
	}
	mq.queue[qo.ID] = qo
	switch index {
	case -1:
		mq.qindex = append(mq.qindex, qo.ID)
	case 0:
		mq.qindex = append([]string{qo.ID}, mq.qindex...)
	default:
		mq.qindex = slices.Insert(mq.qindex, index, qo.ID)
	}
	return qo, nil
}

func (mq *MemoryQueue) PopIndex(index int) (body []byte, object *QueueObject, err error) {
	mq.lock.Lock()
	defer mq.deferFunc()
	if index > len(mq.qindex)-1 {
		return []byte{}, nil, fmt.Errorf("index %v out of bounds", index)
	}
	var ID string
	switch index {
	case -1:
		ID = mq.qindex[:len(mq.qindex)-1][0]
	default:
		ID = mq.qindex[index]
		mq.qindex = append(mq.qindex[:index], mq.qindex[index+1:]...)
	}
	if _, ok := mq.queue[ID]; !ok {
		return []byte{}, nil, fmt.Errorf("item at index %v does not exist in queue", index)
	}
	qo := mq.queue[ID]
	delete(mq.queue, ID)
	return qo.Body, qo, nil
}

func (mq *MemoryQueue) PopID(id string) (body []byte, object *QueueObject, err error) {
	mq.lock.Lock()
	defer mq.deferFunc()
	idx := slices.Index(mq.qindex, id)
	if idx == -1 {
		return []byte{}, nil, fmt.Errorf("item with ID %s does not exist in the queue index", id)
	}
	mq.qindex = append(mq.qindex[:idx], mq.qindex[idx+1:]...)
	if _, ok := mq.queue[id]; !ok {
		return []byte{}, nil, fmt.Errorf("item with ID %s does not exist in the queue", id)
	}
	item := mq.queue[id]
	delete(mq.queue, id)
	return item.Body, item, nil
}

func (mq *MemoryQueue) PeekIndex(index int) (body []byte, object *QueueObject, err error) {
	mq.lock.Lock()
	defer mq.deferFunc()
	if index > len(mq.qindex)-1 {
		return []byte{}, nil, fmt.Errorf("index %v out of bounds", index)
	}
	var ID string
	switch index {
	case -1:
		ID = mq.qindex[len(mq.qindex)-1]
	default:
		ID = mq.qindex[index]
	}
	if _, ok := mq.queue[ID]; !ok {
		return []byte{}, nil, fmt.Errorf("item at index %v does not exist in queue", index)
	}
	qo := mq.queue[ID]
	return qo.Body, qo, nil
}

func (mq *MemoryQueue) PeekID(id string) (body []byte, object *QueueObject, err error) {
	mq.lock.Lock()
	defer mq.deferFunc()
	idx := slices.Index(mq.qindex, id)
	if idx == -1 {
		return []byte{}, nil, fmt.Errorf("item with ID %s does not exist in the queue index", id)
	}
	if _, ok := mq.queue[id]; !ok {
		return []byte{}, nil, fmt.Errorf("item with ID %s does not exist in the queue", id)
	}
	return mq.queue[id].Body, mq.queue[id], nil
}

func (mq *MemoryQueue) PeekScan() (bodies [][]byte, objects map[int]*QueueObject, err error) {
	mq.lock.Lock()
	defer mq.deferFunc()
	for idx, i := range mq.qindex {
		item := mq.queue[i]
		bodies = append(bodies, item.Body)
		objects[idx] = item
	}
	return bodies, objects, nil
}

func (mq *MemoryQueue) Find(match []byte) (found bool, index int, object *QueueObject, err error) {
	mq.lock.Lock()
	defer mq.deferFunc()
	for idx, i := range mq.qindex {
		item := mq.queue[i]
		if bytes.Contains(item.Body, match) {
			return true, idx, item, nil
		}
	}
	return false, -1, nil, nil
}

func (mq *MemoryQueue) ClearQueue() error {
	mq.lock.Lock()
	defer mq.deferFunc()
	mq.qindex = []string{}
	mq.queue = map[string]*QueueObject{}
	return nil
}

func (mq *MemoryQueue) RebuildIndex() error {
	mq.lock.Lock()
	defer mq.deferFunc()
	mq.qindex = []string{}
	for k := range mq.queue {
		mq.qindex = append(mq.qindex, k)
	}
	return nil
}

func (mq *MemoryQueue) DeleteQueue() error {
	mq.lock.Lock()
	defer mq.deferFunc()
	return mq.ClearQueue()
}

func (mq *MemoryQueue) DebugIndex() {
	b, _ := json.MarshalIndent(mq.qindex, "", "    ")
	fmt.Println(string(b[:]))
}

func (mq *MemoryQueue) DebugQueue() {
	b, _ := json.MarshalIndent(mq.queue, "", "    ")
	fmt.Println(string(b[:]))
}

func (mq *MemoryQueue) deferFunc() {
	mq.lock.Unlock()
}
