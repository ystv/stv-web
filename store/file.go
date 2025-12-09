package store

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ystv/stv-web/storage"
)

// FileBackend Applications: apps, Prefix: prefix
type FileBackend struct {
	path  string
	cache *storage.STV
	mutex sync.RWMutex
}

func NewFileBackend(root bool) (Backend, error) {
	var fb *FileBackend

	if root {
		fb = &FileBackend{path: "/db/store.db"}
	} else {
		fb = &FileBackend{path: "./db/store.db"}
	}

	state, err := fb.read(root)
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
func (fb *FileBackend) read(root bool) (*storage.STV, error) {
	var stv storage.STV

	if root {
		_, err := os.Stat("/db")
		if err != nil {
			err = os.Mkdir("/db", 0777)
			if err != nil {
				return nil, fmt.Errorf("failed to make folder /db: %w", err)
			}
		}
	} else {
		_, err := os.Stat("./db")
		if err != nil {
			err = os.Mkdir("./db", 0777)
			if err != nil {
				return nil, fmt.Errorf("failed to make folder ./db: %w", err)
			}
		}
	}

	data, err := os.ReadFile(fb.path)
	// Non-existing stv is ok
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("no previous file read: %w", err)
	}
	if err == nil {
		if err := proto.Unmarshal(data, &stv); err != nil {
			return nil, fmt.Errorf("failed to parse stream stv: %w", err)
		}
	}

	log.Printf("db file from: %s", fb.path)
	return &stv, nil
}

// Save stores the store state in a file
func (fb *FileBackend) save(stv *storage.STV) error {
	out, err := proto.Marshal(stv)
	if err != nil {
		return fmt.Errorf("failed to encode stv: %w", err)
	}
	tmp := fmt.Sprintf(fb.path+".%v", time.Now().Format("2006-01-02T15-04-05"))
	if err := os.WriteFile(tmp, out, 0600); err != nil {
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
