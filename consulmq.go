package consulmq

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

// MQ provides methods for manipulating the message queue
type MQ struct {
	client  *api.Client
	agent   *api.Agent
	kv      *api.KV
	session *api.Session
	qname   string
	id      string
	ip      string
	q       *queue
}

// Config is for passing configuration into the Connect function
type Config struct {
	// Address and port number of the Consul endpoint to connect to
	// EX: 172.16.0.2:8500
	// Default is "localhost:8500"
	Address string `yaml:"address"`
	// Datacenter is a Consul concept that allows for separating assets
	// Consul and ConsulMQ's default is "dc1"
	Datacenter string `yaml:"datacenter"`
	// Consul ACL Token
	// Default is empty (no token)
	Token string `yaml:"token"`
	// Unqiue name of the message queue
	// Default is "consulmq"
	MQName string `yaml:"mqname"`
	// A TTL for messages on the queue
	// Default is 10 years (effectively no TTL)
	// TODO: Enforce TTL's
	TTL time.Duration
}

var defaults = map[string]string{
	"Address":    "localhost:8500",
	"Datacenter": "dc1",
	"MQName":     "consulmq",
}

type queue struct {
	Name       string
	RootPath   string        `json:"root_path"`
	SystemPath string        `json:"system_path"`
	QueuePath  string        `json:"queue_path"`
	RetryPath  string        `json:"retry_path"`
	CreatedAt  time.Time     `json:"created_at"`
	TTL        time.Duration `json:"ttl"`
}

// QueueObject is a container around any data in the queue
type QueueObject struct {
	// Unique ID of the object
	ID string
	// Creation time of the object
	CreatedAt time.Time
	// When the object will be deleted
	TTLDeadline time.Time
	// Any tags for the object (TBI)
	// TODO: Implement message tagging
	Tags []string
	// The actual data to be put on the queue
	Body []byte
}

// Connect sets up the connection to the message queue
// Connect will generate a unique machine ID that persists across restarts and will
// use this ID to register as a service with Consul.
func Connect(config Config) (*MQ, error) {
	c := api.DefaultConfig()
	config = setDefaults(config, defaults)
	c.Address = config.Address
	c.Datacenter = config.Datacenter
	c.Token = config.Token
	client, err := api.NewClient(c)
	if err != nil {
		return nil, err
	}
	id, err := machineid.ID()
	if err != nil {
		return nil, err
	}
	ip, err := getIP(config.Address)
	if err != nil {
		return nil, err
	}
	mq := &MQ{
		client:  client,
		agent:   client.Agent(),
		kv:      client.KV(),
		session: client.Session(),
		qname:   "consulmq/" + config.MQName,
		id:      id,
		ip:      ip,
	}
	q, err := mq.getQueueInfo(config)
	mq.q = &q
	err = registerServiceConsul(mq)
	if err != nil {
		return nil, err
	}
	go mq.doTTLUpdate(false)
	if err != nil {
		return nil, err
	}
	err = mq.createPaths()
	if err != nil {
		return nil, err
	}
	return mq, nil
}

func (mq *MQ) doTTLUpdate(once bool) error {
	ticker := time.NewTicker(1 * time.Second)
	var errs error
	for {
		select {
		case <-ticker.C:
			err := mq.agent.UpdateTTL("service:consulmq-"+mq.id, "OK", "passing")
			if err != nil {
				errs = err
			}
		}
		if once {
			return errs
		}
	}
}

func (mq *MQ) getQueueInfo(config Config) (queue, error) {
	obj, _, err := mq.kv.Get(mq.qname+"/_system/info", nil)
	if err != nil {
		return queue{}, err
	}
	if obj == nil {
		return mq.makeQueueInfo(config)
	}
	var info queue
	err = json.Unmarshal(obj.Value, &info)
	if err != nil {
		return queue{}, nil
	}
	return info, nil
}

func (mq *MQ) makeQueueInfo(config Config) (queue, error) {
	q := &queue{
		Name:       config.MQName,
		RootPath:   mq.qname + "/",
		SystemPath: mq.qname + "/_system/",
		QueuePath:  mq.qname + "/q/",
		CreatedAt:  time.Now(),
		TTL:        config.TTL,
	}
	b, err := json.MarshalIndent(q, "", "    ")
	if err != nil {
		return queue{}, nil
	}
	_, err = mq.kv.Put(&api.KVPair{
		Key:   q.SystemPath + "info",
		Value: b,
	}, nil)
	if err != nil {
		return queue{}, nil
	}
	return *q, nil
}

func (mq *MQ) createPaths() error {
	for _, p := range []string{mq.q.QueuePath} {
		obj, _, err := mq.kv.Get(p+"_index", nil)
		if err != nil {
			return err
		}
		if obj == nil {
			b, err := json.MarshalIndent([]string{}, "", "    ")
			if err != nil {
				return err
			}
			_, err = mq.kv.Put(&api.KVPair{
				Key:   p + "_index",
				Value: b,
			}, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (mq *MQ) loadIndex(q string) ([]string, *api.KVPair, error) {
	obj, _, err := mq.kv.Get(mq.q.RootPath+q+"/_index", nil)
	if err != nil || obj == nil {
		return nil, obj, err
	}
	var idx []string
	err = json.Unmarshal(obj.Value, &idx)
	if err != nil {
		return nil, obj, err
	}
	return idx, obj, nil
}

func (mq *MQ) writeIndex(q string, idx []string, kv *api.KVPair) error {
	sess, _, err := mq.session.CreateNoChecks(nil, nil)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(idx, "", "    ")
	if err != nil {
		return err
	}
	k := &api.KVPair{
		Key:     mq.q.RootPath + q + "/_index",
		Value:   b,
		Session: sess,
	}
	lock := false
	kv.Session = sess
	for i := 0; i <= 10; i++ {
		locked, _, err := mq.kv.Acquire(kv, nil)
		if err != nil {
			return err
		}
		if locked {
			lock = locked
		}
	}
	defer mq.unlock(k)
	if !lock {
		return fmt.Errorf("unable to acquire index lock for queue " + q)
	}

	_, err = mq.kv.Put(k, nil)
	return nil
}

func (mq *MQ) unlock(kv *api.KVPair) {
	unlock, _, err := mq.kv.Release(kv, nil)
	if err != nil {
		panic(err)
	}
	if !unlock {
		panic(fmt.Errorf("unable to release lock on index " + kv.Key))
	}
}

func (mq *MQ) indexPush(queue string, id string) error {
	idx, kv, err := mq.loadIndex(queue)
	if err != nil {
		return err
	}
	idx = append(idx, id)
	err = mq.writeIndex(queue, idx, kv)
	if err != nil {
		return err
	}
	return nil
}

func (mq *MQ) indexPushFirst(queue string, id string) error {
	idx, kv, err := mq.loadIndex(queue)
	if err != nil {
		return err
	}
	idx = append([]string{id}, idx...)
	err = mq.writeIndex(queue, idx, kv)
	if err != nil {
		return err
	}
	return nil
}

func (mq *MQ) indexPop(queue string) (string, int, error) {
	idx, kv, err := mq.loadIndex(queue)
	if err != nil {
		return "", len(idx), err
	}
	var id string
	if len(idx) > 0 {
		id, idx = idx[0], idx[1:]
		err = mq.writeIndex(queue, idx, kv)
		if err != nil {
			return "", len(idx), err
		}
	}
	return id, len(idx), nil
}

func (mq *MQ) indexPopLast(queue string) (string, int, error) {
	idx, kv, err := mq.loadIndex(queue)
	if err != nil {
		return "", len(idx), err
	}
	var id string
	if len(idx) > 0 {
		id = idx[len(idx)-1]
		idx[len(idx)-1] = ""
		idx = idx[:len(idx)-1]
		err = mq.writeIndex(queue, idx, kv)
		if err != nil {
			return "", len(idx), err
		}
	}
	return id, len(idx), nil

}

// Push an object to the rear of the queue. Push returns a QueueObject with the object ID,
// the object's CTime, the TTL deadline, and the body that was passed to the function
func (mq *MQ) Push(body []byte) (*QueueObject, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return &QueueObject{}, err
	}
	obj := &QueueObject{
		ID:          id.String(),
		CreatedAt:   time.Now(),
		TTLDeadline: time.Now().Add(mq.q.TTL),
		Body:        body,
	}
	b, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return &QueueObject{}, nil
	}
	err = mq.indexPush("q", id.String())
	if err != nil {
		return &QueueObject{}, err
	}
	_, err = mq.kv.Put(&api.KVPair{
		Key:   mq.q.QueuePath + id.String(),
		Value: b,
	}, nil)
	if err != nil {
		return obj, err
	}
	return obj, nil
}

//PushFirst pushes a new element to the front of the queue. PushFirst returns a QueueObject with the object ID,
// the object's CTime, the TTL deadline, and the body that was passed to the function
func (mq *MQ) PushFirst(body []byte) (*QueueObject, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return &QueueObject{}, err
	}
	obj := &QueueObject{
		ID:          id.String(),
		CreatedAt:   time.Now(),
		TTLDeadline: time.Now().Add(mq.q.TTL),
		Body:        body,
	}
	b, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return &QueueObject{}, nil
	}
	err = mq.indexPushFirst("q", id.String())
	if err != nil {
		return &QueueObject{}, err
	}
	_, err = mq.kv.Put(&api.KVPair{
		Key:   mq.q.QueuePath + id.String(),
		Value: b,
	}, nil)
	if err != nil {
		return obj, err
	}
	return obj, nil
}

// Pop removes the next object in the queue. Pop returns the message body as bytes and a QueueObject with the object ID,
// the object's CTime, the TTL deadline, and the message body as bytes.
func (mq *MQ) Pop() ([]byte, *QueueObject, error) {
	id, _, err := mq.indexPop("q")
	if err != nil {
		return []byte{}, &QueueObject{}, err
	}
	obj, _, err := mq.kv.Get(mq.q.QueuePath+id, nil)
	if err != nil {
		return []byte{}, &QueueObject{}, err
	}
	if obj == nil {
		return []byte{}, &QueueObject{}, fmt.Errorf("object at head is nil")
	}
	_, err = mq.kv.Delete(mq.q.QueuePath+id, nil)
	if err != nil {
		return []byte{}, &QueueObject{}, err
	}
	var qo QueueObject
	err = json.Unmarshal(obj.Value, &qo)
	if err != nil {
		return []byte{}, &QueueObject{}, err
	}
	return qo.Body, &qo, nil
}

// PopLast removes the last (newest) object in the queue. PopLast returns the message body as bytes and a QueueObject with the object ID,
// the object's CTime, the TTL deadline, and the message body as bytes.
func (mq *MQ) PopLast() ([]byte, *QueueObject, error) {
	id, _, err := mq.indexPopLast("q")
	if err != nil {
		return []byte{}, &QueueObject{}, err
	}
	obj, _, err := mq.kv.Get(mq.q.QueuePath+id, nil)
	if err != nil {
		return []byte{}, &QueueObject{}, err
	}
	if obj == nil {
		return []byte{}, &QueueObject{}, fmt.Errorf("object at tail is nil")
	}
	_, err = mq.kv.Delete(mq.q.QueuePath+id, nil)
	if err != nil {
		return []byte{}, &QueueObject{}, err
	}
	var qo QueueObject
	err = json.Unmarshal(obj.Value, &qo)
	if err != nil {
		return []byte{}, &QueueObject{}, err
	}
	return qo.Body, &qo, nil
}
