// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package contextdb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	pkgerrors "github.com/pkg/errors"
)

type MockConDb struct {
	Items []sync.Map
	sync.Mutex
	Err error
}

func (c *MockConDb) Put(ctx context.Context, key string, value interface{}) error {

	var vg interface{}
	err := c.Get(ctx, key, interface{}(&vg))
	if vg != "" {
		c.Delete(ctx, key)
	}
	v, err := json.Marshal(value)
	if err != nil {
		fmt.Println("Error during json marshal")
	}
	var d sync.Map
	d.Store(key, v)
	c.Lock()
	defer c.Unlock()
	c.Items = append(c.Items, d)
	return c.Err
}
func (c *MockConDb) PutWithCheck(ctx context.Context, key string, value interface{}) error {

	return c.Put(ctx, key, value)
}
func (c *MockConDb) HealthCheck() error {
	return c.Err
}
func (c *MockConDb) Get(ctx context.Context, key string, value interface{}) error {
	c.Lock()
	defer c.Unlock()
	for _, item := range c.Items {
		d := make(map[string][]byte)
		item.Range(func(k, v interface{}) bool {
			d[fmt.Sprint(k)] = v.([]byte)
			return true
		})
		for k, v := range d {
			if k == key {
				err := json.Unmarshal([]byte(v), value)
				if err != nil {
					fmt.Println("Error during json unmarshal", err, key)
				}
				return c.Err
			}
		}
	}

	value = nil
	return pkgerrors.Errorf("Key doesn't exist")
}
func (c *MockConDb) GetAllKeys(ctx context.Context, path string) ([]string, error) {
	c.Lock()
	defer c.Unlock()
	n := 0
	for _, item := range c.Items {
		d := make(map[string][]byte)
		item.Range(func(k, v interface{}) bool {
			d[fmt.Sprint(k)] = v.([]byte)
			return true
		})
		for k, _ := range d {
			ok := strings.HasPrefix(k, path)
			if ok {
				n++
			}
		}
	}
	if n == 0 {
		return nil, c.Err
	}

	retk := make([]string, n)

	i := 0
	for _, item := range c.Items {
		d := make(map[string][]byte)
		item.Range(func(k, v interface{}) bool {
			d[fmt.Sprint(k)] = v.([]byte)
			return true
		})
		for k, _ := range d {
			ok := strings.HasPrefix(k, path)
			if ok {
				retk[i] = k
				i++
			}
		}
	}
	return retk, c.Err
}
func (c *MockConDb) Delete(ctx context.Context, key string) error {
	c.Lock()
	defer c.Unlock()
	for i, item := range c.Items {
		d := make(map[string][]byte)
		item.Range(func(k, v interface{}) bool {
			d[fmt.Sprint(k)] = v.([]byte)
			return true
		})
		for k, _ := range d {
			if k == key {
				c.Items[i] = c.Items[len(c.Items)-1]
				c.Items = c.Items[:len(c.Items)-1]
				return c.Err
			}
		}
	}
	return c.Err
}
func (c *MockConDb) DeleteAll(ctx context.Context, key string) error {
	c.Lock()
	defer c.Unlock()
	for i, item := range c.Items {
		d := make(map[string][]byte)
		item.Range(func(k, v interface{}) bool {
			d[fmt.Sprint(k)] = v.([]byte)
			return true
		})
		for k, _ := range d {
			ok := strings.HasPrefix(k, key)
			if ok {
				c.Items[i] = c.Items[len(c.Items)-1]
				c.Items = c.Items[:len(c.Items)-1]
			}
		}
	}
	return c.Err
}
