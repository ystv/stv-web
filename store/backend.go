package store

import (
	"github.com/ystv/stv-web/storage"
)

type Backend interface {
	Read() (*storage.STV, error)
	Write(state *storage.STV) error
}
