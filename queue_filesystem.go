package kvmq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

type FilesystemQueue struct {
	Name        string
	basepath    string
	indexFile   string
	lockTimeout time.Duration
	lockTTL     time.Duration
	qindex      []string
}

func NewFilesystemQueue(config *Config) (*FilesystemQueue, error) {
	if !strings.HasSuffix(config.FilesystemConfig.Directory, "/") {
		config.FilesystemConfig.Directory = config.FilesystemConfig.Directory + "/"
	}
	f := &FilesystemQueue{
		basepath:    config.FilesystemConfig.Directory,
		lockTimeout: time.Duration(config.LockTimeout) * time.Second,
		lockTTL:     time.Duration(config.LockTTL) * time.Second,
	}
	return f, nil
}

func (mq *FilesystemQueue) Connect() error {
	if _, err := os.Stat(mq.basepath + mq.indexFile); errors.Is(err, os.ErrNotExist) {
		mq.qindex = []string{}
		err := mq.writeIndexFile()
		if err != nil {
			return err
		}
	}
	err := mq.readIndexFile()
	if err != nil {
		return err
	}
	return nil
}

func (mq *FilesystemQueue) Length() (int, error) {
	err := mq.readIndexFile()
	if err != nil {
		return -1, err
	}
	return len(mq.qindex), nil
}

func (mq *FilesystemQueue) PushIndex(body []byte, index int) (object *QueueObject, err error) {
	err = mq.lock()
	if err != nil {
		return nil, err
	}
	defer mq.unlock()
	err = mq.readIndexFile()
	if err != nil {
		return nil, err
	}
	if index > len(mq.qindex)-1 {
		return nil, fmt.Errorf("index %v out of bounds", index)
	}
	item := &QueueObject{
		ID:        hash(body),
		CreatedAt: time.Now(),
		Body:      body,
	}
	err = mq.writeObjectFile(item)
	if err != nil {
		return item, err
	}
	switch index {
	case -1:
		mq.qindex = append(mq.qindex, item.ID)
	case 0:
		mq.qindex = append([]string{item.ID}, mq.qindex...)
	default:
		mq.qindex = slices.Insert(mq.qindex, index, item.ID)
	}
	return item, nil
}

func (mq *FilesystemQueue) PopIndex(index int) (body []byte, object *QueueObject, err error) {
	err = mq.lock()
	if err != nil {
		return nil, nil, err
	}
	defer mq.unlock()
	err = mq.readIndexFile()
	if err != nil {
		return nil, nil, err
	}
	if index > len(mq.qindex)-1 {
		return []byte{}, nil, fmt.Errorf("index %v out of bounds", index)
	}
	var id string
	switch index {
	case -1:
		id = mq.qindex[:len(mq.qindex)-1][0]
	default:
		id = mq.qindex[index]
		mq.qindex = append(mq.qindex[:index], mq.qindex[index+1:]...)
	}
	item, err := mq.readObjectFile(id, true)
	if err != nil {
		return nil, nil, err
	}
	return item.Body, item, nil
}

func (mq *FilesystemQueue) PopID(id string) (body []byte, object *QueueObject, err error) {
	err = mq.lock()
	if err != nil {
		return nil, nil, err
	}
	defer mq.unlock()
	err = mq.readIndexFile()
	if err != nil {
		return nil, nil, err
	}
	idx := slices.Index(mq.qindex, id)
	if idx == -1 {
		return []byte{}, nil, fmt.Errorf("item with ID %s does not exist in the queue index", id)
	}
	mq.qindex = append(mq.qindex[:idx], mq.qindex[idx+1:]...)
	item, err := mq.readObjectFile(id, true)
	if err != nil {
		return nil, nil, err
	}
	return item.Body, item, nil
}

func (mq *FilesystemQueue) PeekIndex(index int) (body []byte, object *QueueObject, err error) {
	err = mq.lock()
	if err != nil {
		return nil, nil, err
	}
	defer mq.unlock()
	err = mq.readIndexFile()
	if err != nil {
		return nil, nil, err
	}
	if index > len(mq.qindex)-1 {
		return []byte{}, nil, fmt.Errorf("index %v out of bounds", index)
	}
	var id string
	switch index {
	case -1:
		id = mq.qindex[len(mq.qindex)-1]
	default:
		id = mq.qindex[index]
	}
	item, err := mq.readObjectFile(id, false)
	if err != nil {
		return nil, nil, err
	}
	return item.Body, item, nil
}

func (mq *FilesystemQueue) PeekID(id string) (body []byte, object *QueueObject, err error) {
	err = mq.lock()
	if err != nil {
		return nil, nil, err
	}
	defer mq.unlock()
	err = mq.readIndexFile()
	if err != nil {
		return nil, nil, err
	}
	idx := slices.Index(mq.qindex, id)
	if idx == -1 {
		return []byte{}, nil, fmt.Errorf("item with ID %s does not exist in the queue index", id)
	}
	item, err := mq.readObjectFile(id, false)
	if err != nil {
		return nil, nil, err
	}
	return item.Body, item, nil
}

func (mq *FilesystemQueue) PeekScan() (bodies [][]byte, objects map[int]*QueueObject, err error) {
	err = mq.lock()
	if err != nil {
		return nil, nil, err
	}
	defer mq.unlock()
	err = mq.readIndexFile()
	if err != nil {
		return nil, nil, err
	}
	for idx, i := range mq.qindex {
		item, err := mq.readObjectFile(i, false)
		if err != nil {
			return nil, nil, err
		}
		bodies = append(bodies, item.Body)
		objects[idx] = item
	}
	return bodies, objects, nil
}

func (mq *FilesystemQueue) Find(match []byte) (found bool, index int, object *QueueObject, err error) {
	err = mq.lock()
	if err != nil {
		return false, -1, nil, err
	}
	defer mq.unlock()
	err = mq.readIndexFile()
	if err != nil {
		return false, -1, nil, err
	}
	for idx, i := range mq.qindex {
		item, err := mq.readObjectFile(i, false)
		if err != nil {
			return false, -1, nil, err
		}
		if bytes.Contains(item.Body, match) {
			return true, idx, item, nil
		}
	}
	return false, -1, nil, nil
}

func (mq *FilesystemQueue) ClearQueue() error {
	err := mq.lock()
	if err != nil {
		return err
	}
	defer mq.unlock()
	err = mq.readIndexFile()
	if err != nil {
		return err
	}
	for _, i := range mq.qindex {
		err = os.Remove(mq.basepath + i)
		if err != nil {
			return err
		}
	}
	mq.qindex = []string{}
	err = mq.writeIndexFile()
	if err != nil {
		return err
	}
	return nil
}

func (mq *FilesystemQueue) RebuildIndex() error {
	err := mq.lock()
	if err != nil {
		return err
	}
	defer mq.unlock()
	return nil
}

func (mq *FilesystemQueue) DeleteQueue() error {
	err := mq.lock()
	if err != nil {
		return err
	}
	defer mq.unlock()
	err = mq.readIndexFile()
	if err != nil {
		return err
	}
	for _, i := range mq.qindex {
		err = os.Remove(mq.basepath + i)
		if err != nil {
			return err
		}
	}
	err = os.Remove(mq.basepath + mq.indexFile)
	if err != nil {
		return err
	}
	mq.qindex = []string{}
	return nil
}

func (mq *FilesystemQueue) DebugIndex() {
	mq.readIndexFile()
	b, _ := json.MarshalIndent(mq.qindex, "", " ")
	fmt.Println(b)
}

func (mq *FilesystemQueue) DebugQueue() {
	for _, i := range mq.qindex {
		obj, _ := mq.readObjectFile(i, false)
		b, _ := json.MarshalIndent(obj, "", " ")
		fmt.Println(b)
	}
}

func (mq *FilesystemQueue) readIndexFile() error {
	var idx []string
	b, err := os.ReadFile(mq.basepath + mq.indexFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &idx)
	if err != nil {
		return err
	}
	copy(mq.qindex, idx)
	return nil
}

func (mq *FilesystemQueue) writeIndexFile() error {
	b, err := json.Marshal(mq.basepath + mq.indexFile)
	if err != nil {
		return err
	}
	err = mq.lock()
	if err != nil {
		return err
	}
	defer mq.unlock()
	err = os.WriteFile(mq.basepath+mq.indexFile, b, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (mq *FilesystemQueue) readObjectFile(id string, consume bool) (*QueueObject, error) {
	var q QueueObject
	b, err := os.ReadFile(mq.basepath + id)
	if err != nil {
		return &q, err
	}
	err = json.Unmarshal(b, &q)
	if err != nil {
		return nil, err
	}
	if consume {
		err = os.Remove(mq.basepath + id)
		if err != nil {
			return &q, err
		}
	}
	return &q, nil
}

func (mq *FilesystemQueue) writeObjectFile(q *QueueObject) error {
	b, err := json.Marshal(q)
	if err != nil {
		return err
	}
	err = mq.lock()
	if err != nil {
		return err
	}
	defer mq.unlock()
	err = os.WriteFile(mq.basepath+q.ID, b, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (mq *FilesystemQueue) lock() error {
	var oldLock Lock
	if b, err := os.ReadFile(mq.basepath + ".lock"); err == nil {
		// lock exists
		err = json.Unmarshal(b, &oldLock)
		if err != nil {
			return err
		}
		// check if TTL is old or if we're the onwer of the lock, if so, remove the lock
		if time.Since(oldLock.TTL) >= 0 || oldLock.PID == os.Getpid() {
			err = os.Remove(mq.basepath + ".lock")
			if err != nil {
				return err
			}
		}
		// if we're unable to modify it, we wait
		elapsed := 0 * time.Millisecond
		for {
			if elapsed >= mq.lockTimeout {
				// timeout
				return fmt.Errorf("QTIMEOUT: unable to acquire queue lock")
			}
			if _, err := os.Stat(mq.basepath + ".lock"); errors.Is(err, os.ErrNotExist) {
				// break if file is removed
				break
			}
			// sleep and update elapsed
			time.Sleep(100 * time.Millisecond)
			elapsed += 100
		}
	}
	l := &Lock{
		PID: os.Getpid(),
		TTL: time.Now().Add(mq.lockTTL),
	}
	byt, err := json.Marshal(l)
	if err != nil {
		return err
	}
	err = os.WriteFile(mq.basepath+".lock", byt, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (mq *FilesystemQueue) unlock() {
	if b, err := os.ReadFile(mq.basepath + ".lock"); err == nil {
		// lock exits
		var oldLock Lock
		_ = json.Unmarshal(b, &oldLock)
		// its exired or its ours, so remove it
		if time.Since(oldLock.TTL) >= 0 || oldLock.PID == os.Getpid() {
			_ = os.Remove(mq.basepath + ".lock")
		}
	}
}
