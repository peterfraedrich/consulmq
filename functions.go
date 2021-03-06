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

func setDefaults(config *Config, newconf *Config, defaults map[string]string) *Config {
	if config.Address == "" {
		newconf.Address = defaults["Address"]
	} else {
		newconf.Address = config.Address
	}
	if config.Datacenter == "" {
		newconf.Datacenter = defaults["Datacenter"]
	} else {
		newconf.Datacenter = config.Datacenter
	}
	if config.MQName == "" {
		newconf.MQName = defaults["MQName"]
	} else {
		newconf.MQName = config.MQName
	}
	if config.TTL == 0*time.Second {
		newconf.TTL = 87600 * time.Hour
	} else {
		newconf.TTL = config.TTL
	}
	newconf.Token = config.Token
	return newconf
}
