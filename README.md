Store
=====

[![Build Status](https://travis-ci.org/kvij/store?branch=master)](https://travis-ci.org/kvij/store)

Simple thread safe store and retrieve library for Go.

Useful when data must be dynamically added (like sessions or transactions) and retrieved by separate goroutines.

## Installation
Use go get to install and update:

```sh
$ go get -u github.com/kvij/store
```

## Usage

### Basic

```go
package main

import (
	"github.com/kvij/store"
	"time"
)

// Create a new Store only use full when shared somehow
var transactions = store.New()

// Main implements a silly API demo example where a map would be better
func main() {
	handle := transactions.Add(time.Now())

	if t, ok := transactions.Get(handle); ok {
		t.(time.Time).Add(time.Hour)
		transactions.Update(handle, t)
	}

	transactions.Delete(handle)
}
```

### Real world example

A real world example can be found in [Threeleg OAuth](https://github.com/kvij/threeleg). 

### Assertion free

The example above only uses Get() in once place. If this is not the case the type assertions can become a nuisance.
When a store is used consistently to store data of a single type there is an easy fix

```go
package stringstore

import "github.com/kvij/store"

type StringStore struct {
	store.Store	
}

// Method override. Note that StringStore does not adhere to store.Store anymore af this addition.
// Name the func GetString() if it is an issue.
func (ss *StringStore) Get(id string) (value string, ok bool) {
	i, ok:= ss.Store.Get(id)
	if ok {
		value = i.(string)
	}
	
	return
}

//Optional
func (ss *StringStore) Add(value string) (id string) {
	return ss.Store.Add(value)
}

func (ss *StringStore) Update(id, value string) error {
	return ss.Store.Update(id, value)
}
```