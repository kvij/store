package store

import (
	"errors"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	New()
}

func TestNewMapStore(t *testing.T) {
	var i interface{} = NewMapStore()
	if _, ok := i.(Store); !ok {
		t.Error("NewMapStore() does not return a valid Store")
	}
}

func TestMapStore_Add(t *testing.T) {
	ms := NewMapStore()
	wantValue := "Some value"
	wantId := ms.Add(wantValue)
	for gotId, gotValue := range ms.ledger {
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
		{"Error", args{"error"}, errUpdateId, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMapStore()
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
		ms := NewMapStore()
		_, gotOk := ms.Get("DOES NOT EXIST")
		if gotOk {
			t.Errorf("MapStore.Get() gotOk = %v, want %v", gotOk, false)
		}
	})
}

func TestMapStore_Update(t *testing.T) {
	ms := NewMapStore()
	id := "TESTID"
	wrongId := "WRONGID"
	ms.ledger[id] = 15
	wantValue := 42
	err := ms.Update(wrongId, wantValue)
	if err == nil {
		t.Errorf("MapStore.Update() err = %v, want %v", err, errors.New("Update failed no such id"))
	}

	err = ms.Update(id, wantValue)
	if err != nil {
		t.Errorf("MapStore.Update() err = %v, want %v", err, nil)
	}

	if gotValue := ms.ledger[id]; wantValue != gotValue {
		t.Errorf("MapStore.Update() gotValue = %v, want %v", gotValue, wantValue)
	}
}

func TestMapStore_Delete(t *testing.T) {
	id := "TESTID"
	ms := NewMapStore()
	ms.ledger[id] = true
	if _, ok := ms.ledger[id]; !ok {
		t.Errorf("MapStore.Delete() setup failed")
	}

	ms.Delete(id)
	if _, ok := ms.ledger[id]; ok {
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
