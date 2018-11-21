package store

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewStore(t *testing.T) {
	var iface interface{} = New()
	if s, ok := iface.(Store); !ok {
		t.Error("New() does not return a valid Store")
	} else {
		s.Close()
	}

	s := New()
	value := 42
	id := s.Add(value)
	retrieved, ok := s.Get(id)
	if !ok {
		t.Error("Id not found")
	}

	if retrieved.(int) != value {
		t.Errorf("Got = %v, want %v", retrieved, value)
	}
}

func TestMapStore_Add(t *testing.T) {
	s := New()
	wantValue := "Some value"
	wantId := s.Add(wantValue)
	for gotId, gotValue := range s.ledger {
		if gotId != wantId {
			t.Errorf("MapStore.Add() = %v, want %v", gotId, wantId)
		}

		if gotValue != wantValue {
			t.Errorf("MapStore.Add() = %v, want %v", gotValue, wantValue)
		}
	}
}

func TestMapStore_Get(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name      string
		args      args
		wantValue interface{}
		wantOk    bool
	}{
		{"Bool", args{"bool"}, true, true},
		{"Int", args{"int"}, 42, true},
		{"Constant", args{"constant"}, UPDATE, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := New()
			defer ms.Close()
			ms.ledger[tt.args.id] = tt.wantValue
			gotValue, gotOk := ms.Get(tt.args.id)
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("MapStore.Get() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotOk != tt.wantOk {
				t.Errorf("MapStore.Get() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}

	t.Run("Missing", func(t *testing.T) {
		ms := New()
		defer ms.Close()
		_, gotOk := ms.Get("DOES NOT EXIST")
		if gotOk {
			t.Errorf("MapStore.Get() gotOk = %v, want %v", gotOk, false)
		}
	})
}

func TestMapStore_Update(t *testing.T) {
	s := New()
	id := "TESTID"
	wrongId := "WRONGID"
	s.ledger[id] = 15
	wantValue := 42
	err := s.Update(wrongId, wantValue)
	if err == nil {
		t.Errorf("MapStore.Update() err = %v, want %v", err, errors.New("Update failed no such id"))
	}

	err = s.Update(id, wantValue)
	if err != nil {
		t.Errorf("MapStore.Update() err = %v, want %v", err, nil)
	}

	if gotValue := s.ledger[id]; wantValue != gotValue {
		t.Errorf("MapStore.Update() gotValue = %v, want %v", gotValue, wantValue)
	}
}

func TestMapStore_Delete(t *testing.T) {
	id := "TESTID"
	s := New()
	s.ledger[id] = true
	if _, ok := s.ledger[id]; !ok {
		t.Errorf("MapStore.Delete() setup failed")
	}

	s.Delete(id)
	if _, ok := s.ledger[id]; ok {
		t.Errorf("MapStore.Delete() failed: %v still exists", id)
	}

}

type ClashDetector map[string]bool

func (cd ClashDetector) Clash(id string) bool {
	if cd[id] {
		return true
	}

	cd[id] = true
	return false
}

func TestNewId(t *testing.T) {
	cd := ClashDetector(make(map[string]bool))
	cc := 0

	for i := 0; i < 20000; i++ {
		if id := NewId(); cd.Clash(id) {
			t.Log("Clash detected: ", id)
			cc++
		}
	}

	if cc > 0 {
		t.Errorf("Repeated calls to NewId() resulted in %d clashes", cc)
	}
}
