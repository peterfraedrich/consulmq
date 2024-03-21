package kvmq

import (
	"encoding/json"
	"fmt"
)

type RedisQueue struct {
	Name  string
	index string
	queue string
}

func NewRedisQueue(config *Config) (*RedisQueue, error) {

	return &RedisQueue{}, nil
}

func (mq *RedisQueue) Connect() error {

	return nil
}

func (mq *RedisQueue) Length() (int, error) {
	return 0, nil
}

func (mq *RedisQueue) PushIndex(body []byte, index int) (object *QueueObject, err error) {
	return nil, nil
}

func (mq *RedisQueue) PopIndex(index int) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *RedisQueue) PeekIndex(index int) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *RedisQueue) PeekScan() (bodies [][]byte, objects []*QueueObject, err error) {
	return [][]byte{}, nil, nil
}

func (mq *RedisQueue) Find(match []byte) (found bool, index int, object *QueueObject, err error) {
	return false, 0, nil, nil
}

func (mq *RedisQueue) ClearQueue() error {
	return nil
}

func (mq *RedisQueue) RebuildIndex() error {
	return nil
}

func (mq *RedisQueue) DeleteQueue() error {
	return nil
}

func (mq *RedisQueue) DebugIndex() {
	b, _ := json.MarshalIndent(mq.index, "", " ")
	fmt.Println(b)
}

func (mq *RedisQueue) DebugQueue() {
	b, _ := json.MarshalIndent(mq.queue, "", " ")
	fmt.Println(b)
}
