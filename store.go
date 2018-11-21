package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

type Store interface {
	Add(value interface{}) (id string)               // Store data at a new handle
	Get(id string) (value interface{}, ok bool)      // Retrieve stored data by id
	Update(id string, value interface{}) (err error) // Replace data stored at id
	Delete(id string)                                // Remove id from store
}

// Tread safe in memory Store implementation.
// Must be created by store.New() or store.NewMapStore()
type MapStore struct {
	ledger map[string]interface{}
	lock   *sync.Mutex
}

// New returns a new Store. Defaults to MapStore as internal type.
func New() Store {
	return NewMapStore()
}

func NewMapStore() *MapStore {
	return &MapStore{
		ledger: make(map[string]interface{}),
		lock:   &sync.Mutex{},
	}
}

func (ms *MapStore) Add(value interface{}) (id string) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	for taken := true; taken; _, taken = ms.ledger[id] {
		id = NewId()
	}

	ms.ledger[id] = value

	return
}

func (ms *MapStore) Get(id string) (value interface{}, ok bool) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	value, ok = ms.ledger[id]
	return
}

var errUpdateId = errors.New("Update failed no such id")

func (ms *MapStore) Update(id string, value interface{}) (err error) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	if _, ok := ms.ledger[id]; !ok {
		return errUpdateId
	}

	ms.ledger[id] = value
	return
}

func (ms *MapStore) Delete(id string) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	delete(ms.ledger, id)
}

// Generates a random string with enough entrophy to avoid clashes
func NewId() string {
	iv := make([]byte, 8)
	rand.Reader.Read(iv)
	return hex.EncodeToString(iv)
}
