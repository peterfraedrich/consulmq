package kvmq

import (
	"bytes"
	"testing"
)

func TestMQ(t *testing.T) {
	mq, err := NewMQ(&Config{})
	if err != nil {
		t.Error(err)
	}
	mq.Push([]byte("Hello world!"))
}

func TestRDBMS(t *testing.T) {
	mq, err := NewMQ(&Config{})
	if err != nil {
		t.Error(err)
	}
	_, err = mq.Push([]byte("Hello world"))
	if err != nil {
		t.Error(err)
	}
	body, _, err := mq.Pop()
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(body, []byte("Hello world")) {
		t.Errorf("returned unexpected byte slice")
	}
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
	for range [5]int{} {
		mq.Push([]byte("Hello world!"))
	}
}
