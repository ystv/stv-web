package store

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ystv/stv_web/storage"
	"google.golang.org/protobuf/proto"
)

// FileBackend Applications: apps, Prefix: prefix
type FileBackend struct {
	path  string
	cache *storage.STV
	mutex sync.RWMutex
}

func NewFileBackend() (Backend, error) {
	fb := &FileBackend{path: "./db/store.db"}
	state, err := fb.read()
	if err != nil {
		return nil, err
	}
	// persist state
	err = fb.save(state)
	if err != nil {
		return nil, err
	}
	fb.cache = state
	return fb, nil
}

// Read parses the store state from a file
func (fb *FileBackend) read() (*storage.STV, error) {
	var stv storage.STV

	data, err := ioutil.ReadFile(fb.path)
	// Non-existing stv is ok
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("no previous file read: %w", err)
	}
	if err == nil {
		if err := proto.Unmarshal(data, &stv); err != nil {
			return nil, fmt.Errorf("failed to parse stream stv: %w", err)
		}
	}

	log.Println("STV restored from", fb.path)
	return &stv, nil
}

// Save stores the store state in a file
func (fb *FileBackend) save(stv *storage.STV) error {
	out, err := proto.Marshal(stv)
	if err != nil {
		return fmt.Errorf("failed to encode stv: %w", err)
	}
	tmp := fmt.Sprintf(fb.path+".%v", time.Now())
	if err := ioutil.WriteFile(tmp, out, 0600); err != nil {
		return fmt.Errorf("failed to write stv: %w", err)
	}
	err = os.Rename(tmp, fb.path)
	if err != nil {
		return fmt.Errorf("failed to move stv: %w", err)
	}
	return nil
}

func (fb *FileBackend) Read() (*storage.STV, error) {
	fb.mutex.RLock()
	defer fb.mutex.RUnlock()
	return fb.cache, nil
}

func (fb *FileBackend) Write(state *storage.STV) error {
	fb.mutex.Lock()
	defer fb.mutex.Unlock()
	fb.cache = state
	return fb.save(state)
}
