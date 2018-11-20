package store

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

const (
	ADD Operation = iota
	GET
	UPDATE
	DELETE
	CLOSE
)

type Store interface {
	Add(value interface{}) (id string)               // Store data at a new handle
	Get(id string) (value interface{}, ok bool)      // Retrieve stored data by id
	Update(id string, value interface{}) (err error) // Replace data stored at id
	Delete(id string)                                // Remove id from store
	Close() error                                    // Allow the Store to free resources
}

// Tread safe in memory Store implementation.
// Must be created by store.New()
type MapStore struct {
	ledger map[string]interface{}
	req    chan *Message
}

type Operation int

// Internal message format for method <-> worker communication
type Message struct {
	operation Operation
	id        string
	value     interface{}
	err       error
	ok        bool
	resp      chan *Message
}

// New returns a new MapStore pointer and starts an internal worker goroutine. Only the
// worker may modify the fields to avoid race conditions.
//
// Close must be called to let the store go out of scope.
func New() *MapStore {
	ms := &MapStore{
		ledger: make(map[string]interface{}),
		req:    make(chan *Message),
	}

	go func(ms *MapStore) {
		for m := range ms.req {
			switch m.operation {
			case ADD:
				var id string
				for taken := true; taken; _, taken = ms.ledger[id] {
					id = NewId()
				}
				ms.ledger[id] = m.value
				m.id = id
				break
			case GET:
				m.value, m.ok = ms.ledger[m.id]
				break
			case UPDATE:
				if _, ok := ms.ledger[m.id]; !ok {
					m.err = errors.New("Update failed no such id")
				} else {
					ms.ledger[m.id] = m.value
				}
				break
			case DELETE:
				delete(ms.ledger, m.id)
			}

			m.resp <- m
		}
	}(ms)

	return ms
}

func (ms *MapStore) Add(value interface{}) (id string) {
	resp := make(chan *Message)
	m := &Message{
		operation: ADD,
		value:     value,
		resp:      resp,
	}

	ms.req <- m
	m = <-resp

	return m.id
}

func (ms *MapStore) Get(id string) (value interface{}, ok bool) {
	resp := make(chan *Message)
	m := &Message{
		operation: GET,
		id:        id,
		resp:      resp,
	}

	ms.req <- m
	m = <-resp

	return m.value, m.ok
}

func (ms *MapStore) Update(id string, value interface{}) (err error) {
	resp := make(chan *Message)
	m := &Message{
		operation: UPDATE,
		id:        id,
		value:     value,
		resp:      resp,
	}

	ms.req <- m
	m = <-resp

	return m.err
}

func (ms *MapStore) Delete(id string) {
	resp := make(chan *Message)
	m := &Message{
		operation: DELETE,
		id:        id,
		resp:      resp,
	}

	ms.req <- m
	<-resp
}

func (ms *MapStore) Close() error {
	close(ms.req)
	return nil
}

// Generates a random string with enough entrophy to avoid clashes
func NewId() string {
	iv := make([]byte, 8)
	rand.Reader.Read(iv)
	return hex.EncodeToString(iv)
}
