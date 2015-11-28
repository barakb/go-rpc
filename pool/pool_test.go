package pool

import (
	"fmt"
	"math/rand"
	"testing"
)

type resource struct {
	name string
	key  string
}

func TestPool(t *testing.T) {

	freeResource := func(resource interface{}) {
	}
	createResource := func(name string) (interface{}, string, error) {
		key := fmt.Sprintf("%d", rand.Int())
		return &resource{name, key}, key, nil
	}
	const name = "foo"

	pool := CreatePool(1, createResource, freeResource)
	r1, key1, err := pool.Get(name)
	if err != nil || r1 == nil {
		t.Errorf("failed to allocate resource, error is %v\n", err)
	}
	res := r1.(*resource)
	if res.name != "foo" {
		t.Errorf("resource has wrong name %v", res)
	}

	if 0 < pool.Len() {
		t.Errorf("pool len should be zero instead %d", pool.Len())
	}
	pool.Return(res, name, key1)

	if 1 != pool.Len() {
		t.Errorf("pool len should be 1 instead %d", pool.Len())
	}

	r2, key2, err := pool.Get(name)
	if err != nil || r2 == nil {
		t.Errorf("failed to allocate resource, error is %v\n", err)
	}
	if key2 != key1 {
		t.Errorf("pool allocated new object %v while resource exists in pool %v\n", key2, key1)
	}

	if 0 < pool.Len() {
		t.Errorf("pool len should be zero instead %d", pool.Len())
	}

	r1, key1, err = pool.Get(name)
	if r1 == r2 {
		t.Errorf("pool should created a new object")
	}

	pool.Return(r1, name, key1)
	pool.Return(r2, name, key2)

	if 1 != pool.Len() {
		t.Errorf("pool len should be 1 instead %d", pool.Len())
	}

	fmt.Printf("returns key %s, resource %v\n", key1, res)
	//	fmt.Printf("pool is %#v\n", pool)
}

var result interface{}

func BenchmarkPool(b *testing.B) {
	freeResource := func(resource interface{}) {
	}
	createResource := func(name string) (interface{}, string, error) {
		key := fmt.Sprintf("%d", rand.Int())
		return &resource{name, key}, key, nil
	}
	const name = "foo"

	pool := CreatePool(10, createResource, freeResource)
	var val interface{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			val, key, _ := pool.Get(name)
			pool.Return(val, name, key)
		}
	})
	result = val

}
