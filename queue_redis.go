package kvmq

type RedisQueue struct {
	Name  string
	index string
	queue string
}

func NewRedisQueue(config Config) (RedisQueue, error) {

	return RedisQueue{}, nil
}

func (mq *RedisQueue) Connect() error {

	return nil
}
