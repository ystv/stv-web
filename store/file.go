package store

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/ystv/stv_web/storage"
	"google.golang.org/protobuf/proto"
)

type FileBackendConfig struct {
	Path string
}

// Applications: apps, Prefix: prefix
type FileBackend struct {
	path  string
	cache storage.STV
	mutex sync.RWMutex
}

func NewFileBackend(config FileBackendConfig) (Backend, error) {
	fb := &FileBackend{path: config.Path}
	state, err := fb.read()
	if err != nil {
		return nil, err
	}
	// persist state
	fb.save(state)
	fb.cache = *state
	return fb, nil
}

// Read parses the store state from a file
func (fb *FileBackend) read() (*storage.STV, error) {
	var state storage.STV

	data, err := ioutil.ReadFile(fb.path)
	// Non-existing state is ok
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("no previous file read: %w", err)
	}
	if err == nil {
		if err := proto.Unmarshal(data, &state); err != nil {
			return nil, fmt.Errorf("failed to parse stream state: %w", err)
		}
	}

	// Clear active information for old streams
	for _, stream := range state.Streams {
		stream.Active = false
	}

	// Generate secret
	if len(state.Secret) == 0 {
		state.Secret = make([]byte, 32)
		rand.Read(state.Secret)
		fb.save(&state)
	}

	log.Println("STV restored from", fb.path)
	return &state, nil
}

// Save stores the store state in a file
func (fb *FileBackend) save(state *storage.STV) error {
	out, err := proto.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to encode state: %w", err)
	}
	tmp := fmt.Sprintf(fb.path+".%v", time.Now())
	if err := ioutil.WriteFile(tmp, out, 0600); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}
	err = os.Rename(tmp, fb.path)
	if err != nil {
		return fmt.Errorf("failed to move state: %w", err)
	}
	return nil
}

func (fb *FileBackend) Read() (storage.STV, error) {
	fb.mutex.RLock()
	defer fb.mutex.RUnlock()
	return fb.cache, nil
}

func (fb *FileBackend) Write(state storage.STV) error {
	fb.mutex.Lock()
	defer fb.mutex.Unlock()
	fb.cache = state
	return fb.save(&state)
}
