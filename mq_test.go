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
