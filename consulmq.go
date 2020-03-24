package consulmq

import (
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

// MQ provides methods for manipulating the message queue
type MQ struct {
	client *api.Client
	agent  *api.Agent
	kv     *api.KV
	q      string
	id     string
}

// Config is for passing configuration into the Connect function
type Config struct {
	Address    string
	Datacenter string
	Token      string
	MQName     string
}

// Connect sets up the connection to the message queue
func Connect(config Config) (*MQ, error) {
	c := api.DefaultConfig()
	c.Address = config.Address
	c.Datacenter = config.Datacenter
	c.Token = config.Token
	client, err := api.NewClient(c)
	if err != nil {
		return nil, err
	}
	id := strings.Split(uuid.New().String(), "-")[0]
	mq := &MQ{
		client: client,
		agent:  client.Agent(),
		kv:     client.KV(),
		q:      "/consulmq/" + config.MQName,
		id:     id,
	}
	registerServiceConsul(mq)
	return mq, nil
}

func registerServiceConsul(mq *MQ) {
	mq.agent.ServiceRegister(&api.AgentServiceRegistration{
		Name: "consulmq",
		ID:   "consulmq-" + mq.id,
		Tags: []string{"consulmq"},
	})
}
