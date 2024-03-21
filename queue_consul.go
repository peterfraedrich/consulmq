package kvmq

type ConsulQueue struct {
	Name     string
	basepath string
	index    string
	queue    string
}

func NewConsulQueue(config Config) (ConsulQueue, error) {

	return ConsulQueue{}, nil
}

func (mq *ConsulQueue) Connect() error {

	return nil
}
