package consulmq

import (
	"io/ioutil"
	"testing"

	"gopkg.in/yaml.v2"
)

var TestBytes = []byte("This is a test, this is only a test")
var config Config
var mq *MQ

func TestMain(m *testing.M) {
	b, err := ioutil.ReadFile("test.config")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		panic(err)
	}
	mq, err = Connect(config)
	if err != nil {
		panic(err)
	}
	m.Run()
}

func TestConnect(t *testing.T) {
	mq, err := Connect(Config{
		Address:    "172.17.0.2:8500",
		Datacenter: "dc1",
		Token:      "",
		MQName:     "testMQ",
	})
	if err != nil || mq == nil {
		t.Error(err)
	}
}

func TestDoTTLUpdate(t *testing.T) {
	err := mq.doTTLUpdate(true)
	if err != nil {
		t.Error(err)
	}
}

func TestGetQueueInfo(t *testing.T) {
	q, err := mq.getQueueInfo(config)
	emptyq := queue{}
	if err != nil {
		t.Error(err)
	}
	if q == emptyq {
		t.Error("Queue is uninitialized")
	}
}

func TestMakeQueueInfo(t *testing.T) {
	q, err := mq.makeQueueInfo(config)
	if err != nil {
		t.Error(err)
	}
	if q.Name != config.MQName ||
		q.RootPath != mq.qname+"/" ||
		q.SystemPath != mq.qname+"/_system/" ||
		q.QueuePath != mq.qname+"/q/" {
		t.Errorf("Queue not initialized correctly!")
	}
}

func TestPush(t *testing.T) {
	_, err := mq.Push(TestBytes)
	if err != nil {
		t.Error(err)
	}
}

func TestPushFirst(t *testing.T) {
	_, err := mq.PushFirst(TestBytes)
	if err != nil {
		t.Error(err)
	}
}

func TestPop(t *testing.T) {
	data, _, err := mq.Pop()
	if err != nil {
		t.Error(err)
	}
	if string(data[:]) != string(TestBytes[:]) {
		t.Error("Data does not match test string!")
	}
}

func TestPopLast(t *testing.T) {
	data, _, err := mq.PopLast()
	if err != nil {
		t.Error(err)
	}
	if string(data[:]) != string(TestBytes[:]) {
		t.Error("Data does not match test string!")
	}
}
