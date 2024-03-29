package kvmq

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ConsulQueue struct {
	Name        string
	basepath    string
	datacenter  string
	lockTimeout time.Duration
	lockTTL     time.Duration
	index       string
	queue       string
}

func NewConsulQueue(config *Config) (*ConsulQueue, error) {
	if !strings.HasSuffix(config.ConsulConfig.KVPath, "/") {
		config.ConsulConfig.KVPath = config.ConsulConfig.KVPath + "/"
	}
	c := &ConsulQueue{
		datacenter:  config.ConsulConfig.Datacenter,
		basepath:    config.ConsulConfig.KVPath,
		lockTimeout: time.Duration(config.LockTimeout) * time.Second,
		lockTTL:     time.Duration(config.LockTTL) * time.Second,
	}
	return c, nil
}

func (mq *ConsulQueue) Connect() error {

	return nil
}

func (mq *ConsulQueue) Length() (int, error) {
	return 0, nil
}

func (mq *ConsulQueue) PushIndex(body []byte, index int) (object *QueueObject, err error) {
	return nil, nil
}

func (mq *ConsulQueue) PopIndex(index int) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *ConsulQueue) PopID(id string) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *ConsulQueue) PeekIndex(index int) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *ConsulQueue) PeekID(id string) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *ConsulQueue) PeekScan() (bodies [][]byte, objects map[int]*QueueObject, err error) {
	return [][]byte{}, nil, nil
}

func (mq *ConsulQueue) Find(match []byte) (found bool, index int, object *QueueObject, err error) {
	return false, 0, nil, nil
}

func (mq *ConsulQueue) ClearQueue() error {
	return nil
}

func (mq *ConsulQueue) RebuildIndex() error {
	return nil
}

func (mq *ConsulQueue) DeleteQueue() error {
	return nil
}

func (mq *ConsulQueue) DebugIndex() {
	b, _ := json.MarshalIndent(mq.index, "", " ")
	fmt.Println(b)
}

func (mq *ConsulQueue) DebugQueue() {
	b, _ := json.MarshalIndent(mq.queue, "", " ")
	fmt.Println(b)
}
