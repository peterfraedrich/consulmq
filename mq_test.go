package kvmq

import (
	"testing"
)

func TestMQ(t *testing.T) {
	mq, err := NewMQ(&Config{})
	if err != nil {
		t.Error(err)
	}
	mq.Push([]byte("Hello world!"))
	mq.Debug()
}

func TestCustom(t *testing.T) {
	mq, err := NewMQ(&Config{
		Backend: "custom",
	},
		&MemoryQueue{},
	)
	if err != nil {
		t.Error(err)
	}
	for range [100]int{} {
		mq.Push([]byte("Hello world!"))
	}
	mq.Debug()
}
