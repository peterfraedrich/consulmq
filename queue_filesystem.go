package kvmq

type FilesystemQueue struct {
	Name     string
	basepath string
	index    string
	queue    string
}

func NewFilesystemQueue(config Config) (FilesystemQueue, error) {

	return FilesystemQueue{}, nil
}

func (mq *FilesystemQueue) Connect() error {

	return nil
}
