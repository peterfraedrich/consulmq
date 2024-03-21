package kvmq

type MemoryQueue struct {
	Name  string
	index []string
	queue map[string]QueueObject
}

func newMemoryQueue(config Config) (MemoryQueue, error) {

	return MemoryQueue{}, nil
}

func (mq *MemoryQueue) Connect() error {
	mq.index = []string{}
	mq.queue = map[string]QueueObject{}
	return nil
}
