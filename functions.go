package consulmq

import (
	"net"
	"time"

	"github.com/hashicorp/consul/api"
)

func registerServiceConsul(mq *MQ) error {
	err := mq.agent.ServiceRegister(&api.AgentServiceRegistration{
		Name:    "consulmq",
		ID:      "consulmq-" + mq.id,
		Address: mq.ip,
		Tags:    []string{"consulmq"},
		Check: &api.AgentServiceCheck{
			TTL:                            "10s",
			DeregisterCriticalServiceAfter: "1m",
			Status:                         "passing",
		},
	})
	return err
}

func getIP(addr string) (string, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	return localAddr.IP.String(), nil
}

func setDefaults(config Config, defaults map[string]string) Config {
	if config.Address == "" {
		config.Address = defaults["Address"]
	}
	if config.Datacenter == "" {
		config.Datacenter = defaults["Datacenter"]
	}
	if config.MQName == "" {
		config.MQName = defaults["MQName"]
	}
	if config.TTL == 0*time.Second {
		config.TTL = 87600 * time.Hour
	}
	return config
}
