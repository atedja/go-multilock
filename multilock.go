/*
Copyright 2017 Albert Tedja

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package multilock

import (
	"runtime"
	"sort"
)

var locks = struct {
	lock chan byte
	list map[string]chan byte
}{
	lock: make(chan byte, 1),
	list: make(map[string]chan byte),
}

type MultiLock struct {
	keys   []string
	chans  []chan byte
	lock   chan byte
	unlock chan byte
}

func (self *MultiLock) Lock() {
	self.lock <- 1

	// get the channels and attempt to acquire them
	self.chans = make([]chan byte, 0, len(self.keys))
	for i := 0; i < len(self.keys); {
		ch := getChan(self.keys[i])
		_, ok := <-ch
		if ok {
			self.chans = append(self.chans, ch)
			i++
		}
	}

	self.unlock <- 1
}

// Unlocks this lock. Must be called after Lock.
// Can only be invoked if there is a previous call to Lock.
func (self *MultiLock) Unlock() {
	<-self.unlock

	if self.chans != nil {
		for _, ch := range self.chans {
			ch <- 1
		}
		self.chans = nil
	}

	<-self.lock
}

// Temporarily unlocks, gives up the cpu time to other goroutine, and attempts to lock again.
func (self *MultiLock) Yield() {
	self.Unlock()
	runtime.Gosched()
	self.Lock()
}

// Creates a new multilock for the specified keys
func New(locks ...string) *MultiLock {
	if len(locks) == 0 {
		return nil
	}

	locks = unique(locks)
	sort.Strings(locks)
	return &MultiLock{
		keys:   locks,
		lock:   make(chan byte, 1),
		unlock: make(chan byte, 1),
	}
}

// Attempts to lock multiple keys. Keys must be unique.
// Return an MultiLock instance that must be unlocked.
func Lock(locks ...string) *MultiLock {
	ml := New(locks...)
	ml.Lock()
	return ml
}

// Unlocks an acquired lock. Must be called after Lock.
func Unlock(ml *MultiLock) {
	if ml == nil {
		return
	}

	ml.Unlock()
}

// Cleans old unused locks. Returns removed keys.
func Clean() []string {
	locks.lock <- 1
	defer func() { <-locks.lock }()

	toDelete := make([]string, 0, len(locks.list))
	for key, ch := range locks.list {
		select {
		case <-ch:
			close(ch)
			toDelete = append(toDelete, key)
		default:
		}
	}

	for _, del := range toDelete {
		delete(locks.list, del)
	}

	return toDelete
}

// Create and get the channel for the specified key.
func getChan(key string) chan byte {
	locks.lock <- 1
	defer func() { <-locks.lock }()

	if locks.list[key] == nil {
		locks.list[key] = make(chan byte, 1)
		locks.list[key] <- 1
	}
	return locks.list[key]
}

// Return a new string with unique elements.
func unique(arr []string) []string {
	if arr == nil || len(arr) <= 1 {
		return arr
	}

	found := map[string]bool{}
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if !found[v] {
			found[v] = true
			result = append(result, v)
		}
	}
	return result
}
