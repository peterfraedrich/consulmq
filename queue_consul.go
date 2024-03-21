package kvmq

import (
	"encoding/json"
	"fmt"
)

type ConsulQueue struct {
	Name     string
	basepath string
	index    string
	queue    string
}

func NewConsulQueue(config *Config) (*ConsulQueue, error) {

	return &ConsulQueue{}, nil
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

func (mq *ConsulQueue) PeekIndex(index int) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *ConsulQueue) PeekScan() (bodies [][]byte, objects []*QueueObject, err error) {
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
