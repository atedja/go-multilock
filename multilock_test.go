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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	fmt.Println() // just so we don't have to remove/unremove fmt
}

func TestUnique(t *testing.T) {
	var arr []string
	assert := assert.New(t)

	arr = []string{"a", "b", "c"}
	assert.Equal(arr, unique(arr))

	arr = []string{"a", "a", "a"}
	assert.Equal([]string{"a"}, unique(arr))

	arr = []string{"a", "a", "b"}
	assert.Equal([]string{"a", "b"}, unique(arr))

	arr = []string{"a", "b", "a"}
	assert.Equal([]string{"a", "b"}, unique(arr))

	arr = []string{"a", "b", "c", "b", "d"}
	assert.Equal([]string{"a", "b", "c", "d"}, unique(arr))
}

func TestGetChan(t *testing.T) {
	ch1 := getChan("a")
	ch2 := getChan("aa")
	ch3 := getChan("a")

	assert := assert.New(t)
	assert.NotEqual(ch1, ch2)
	assert.Equal(ch1, ch3)
}

func TestLockUnlock(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(5)

	go func() {
		lock := Lock("dog", "cat", "owl")
		defer Unlock(lock)

		<-time.After(100 * time.Millisecond)
		wg.Done()
	}()

	go func() {
		lock := Lock("cat", "dog", "bird")
		defer Unlock(lock)

		<-time.After(100 * time.Millisecond)
		wg.Done()
	}()

	go func() {
		lock := Lock("cat", "bird", "owl")
		defer Unlock(lock)

		<-time.After(100 * time.Millisecond)
		wg.Done()
	}()

	go func() {
		lock := Lock("bird", "owl", "snake")
		defer Unlock(lock)

		<-time.After(100 * time.Millisecond)
		wg.Done()
	}()

	go func() {
		lock := Lock("owl", "snake")
		defer Unlock(lock)

		<-time.After(1 * time.Second)
		wg.Done()
	}()

	wg.Wait()
}

func TestYield(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)
	var resources = map[string]int{}

	go func() {
		lock := Lock("A", "C")
		defer Unlock(lock)

		for resources["ac"] == 0 {
			lock.Yield()
		}
		resources["dc"] = 10

		wg.Done()
	}()

	go func() {
		lock := Lock("D", "C")
		defer Unlock(lock)

		resources["ac"] = 5
		for resources["dc"] == 0 {
			lock.Yield()
		}

		wg.Done()
	}()

	wg.Wait()

	assert.Equal(t, 5, resources["ac"])
	assert.Equal(t, 10, resources["dc"])
}

func TestClean(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(3)

	// some goroutine that holds multiple locks
	go1done := make(chan bool, 1)
	go func() {
	Loop:
		for {
			select {
			case <-go1done:
				break Loop
			default:
				lock := Lock("A", "B", "C", "E", "Z")
				<-time.After(30 * time.Millisecond)
				Unlock(lock)
			}
		}
		wg.Done()
	}()

	// another goroutine
	go2done := make(chan bool, 1)
	go func() {
	Loop:
		for {
			select {
			case <-go2done:
				break Loop
			default:
				lock := Lock("B", "C", "K", "L", "Z")
				<-time.After(200 * time.Millisecond)
				Unlock(lock)
			}
		}
		wg.Done()
	}()

	// this one cleans up the locks every 100 ms
	done := make(chan bool, 1)
	go func() {
		c := time.Tick(100 * time.Millisecond)
	Loop:
		for {
			select {
			case <-done:
				break Loop
			case <-c:
				Clean()
			}
		}
		wg.Done()
	}()

	<-time.After(2 * time.Second)
	go1done <- true
	go2done <- true
	<-time.After(1 * time.Second)
	done <- true
	wg.Wait()
	assert.Equal(t, []string{}, Clean())
}

func TestSyncCondCompatibility(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	cond := sync.NewCond(New("A", "C"))

	sharedRsc := "foo"

	go func() {
		cond.L.Lock()
		for sharedRsc == "foo" {
			cond.Wait()
		}
		sharedRsc = "fizz!"
		cond.Broadcast()
		cond.L.Unlock()
		wg.Done()
	}()

	go func() {
		cond.L.Lock()
		sharedRsc = "bar"
		cond.Broadcast()
		for sharedRsc == "bar" {
			cond.Wait()
		}
		cond.L.Unlock()
		wg.Done()
	}()

	wg.Wait()
	assert.Equal(t, "fizz!", sharedRsc)
}
