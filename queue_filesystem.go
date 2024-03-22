package kvmq

import (
	"encoding/json"
	"fmt"
)

type FilesystemQueue struct {
	Name     string
	basepath string
	index    string
	queue    string
}

func NewFilesystemQueue(config *Config) (*FilesystemQueue, error) {

	return &FilesystemQueue{}, nil
}

func (mq *FilesystemQueue) Connect() error {

	return nil
}

func (mq *FilesystemQueue) Length() (int, error) {
	return 0, nil
}

func (mq *FilesystemQueue) PushIndex(body []byte, index int) (object *QueueObject, err error) {
	return nil, nil
}

func (mq *FilesystemQueue) PopIndex(index int) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *FilesystemQueue) PopID(id string) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *FilesystemQueue) PeekIndex(index int) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *FilesystemQueue) PeekID(id string) (body []byte, object *QueueObject, err error) {
	return []byte{}, nil, nil
}

func (mq *FilesystemQueue) PeekScan() (bodies [][]byte, objects []*QueueObject, err error) {
	return [][]byte{}, nil, nil
}

func (mq *FilesystemQueue) Find(match []byte) (found bool, index int, object *QueueObject, err error) {
	return false, 0, nil, nil
}

func (mq *FilesystemQueue) ClearQueue() error {
	return nil
}

func (mq *FilesystemQueue) RebuildIndex() error {
	return nil
}

func (mq *FilesystemQueue) DeleteQueue() error {
	return nil
}

func (mq *FilesystemQueue) DebugIndex() {
	b, _ := json.MarshalIndent(mq.index, "", " ")
	fmt.Println(b)
}

func (mq *FilesystemQueue) DebugQueue() {
	b, _ := json.MarshalIndent(mq.queue, "", " ")
	fmt.Println(b)
}
