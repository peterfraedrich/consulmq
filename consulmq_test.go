package consulmq

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

var config Config
var mq *MQ

func TestMain(m *testing.M) {
	b, err := ioutil.ReadFile("testdata/test.config")
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
	exitVal := m.Run()
	_, err = mq.kv.DeleteTree("consulmq/consulmq/", nil)
	if err != nil {
		panic(err)
	}
	os.Exit(exitVal)
}

func TestConnect(t *testing.T) {
	mq, err := Connect(config)
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

func TestPushPop(t *testing.T) {
	testData := []byte(uuid.New().String())
	err := mq.EmptyQueue()
	if err != nil {
		t.Error(err)
	}
	_, err = mq.Push(testData)
	if err != nil {
		t.Error(err)
	}
	b, _, err := mq.Pop()
	if err != nil {
		t.Error(err)
	}
	if string(testData[:]) != string(b[:]) {
		t.Errorf("Test strings do not match! Had %s got %s", string(testData[:]), string(b[:]))
	}
	err = mq.EmptyQueue()
	if err != nil {
		t.Error(err)
	}
}

func TestPushPeekQueue(t *testing.T) {
	testData := []byte(uuid.New().String())
	err := mq.EmptyQueue()
	if err != nil {
		t.Error(err)
	}
	_, err = mq.Push(testData)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 10; i++ {
		b, _, err := mq.Peek()
		if err != nil {
			t.Error(err)
		}
		if string(testData[:]) != string(b[:]) {
			t.Errorf("Test strings do not match! Had %s got %s", string(testData[:]), string(b[:]))
		}
	}

	err = mq.EmptyQueue()
	if err != nil {
		t.Error(err)
	}
}

func TestPushFirstPopLast(t *testing.T) {
	testData := []byte(uuid.New().String())
	err := mq.EmptyQueue()
	if err != nil {
		t.Error(err)
	}
	_, err = mq.PushFirst(testData)
	if err != nil {
		t.Error(err)
	}
	b, _, err := mq.PopLast()
	if err != nil {
		t.Error(err)
	}
	if string(testData[:]) != string(b[:]) {
		t.Errorf("Test strings do not match! Had %s got %s", string(testData[:]), string(b[:]))
	}
	err = mq.EmptyQueue()
	if err != nil {
		t.Error(err)
	}
}

func TestEmptyQueue(t *testing.T) {
	err := mq.EmptyQueue()
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteQueue(t *testing.T) {
	err := mq.DeleteQueue()
	if err != nil {
		t.Error(err)
	}
}
