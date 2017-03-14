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
		lock := New("dog", "cat", "owl")
		lock.Lock()
		defer lock.Unlock()

		<-time.After(100 * time.Millisecond)
		wg.Done()
	}()

	go func() {
		lock := New("cat", "dog", "bird")
		lock.Lock()
		defer lock.Unlock()

		<-time.After(100 * time.Millisecond)
		wg.Done()
	}()

	go func() {
		lock := New("cat", "bird", "owl")
		lock.Lock()
		defer lock.Unlock()

		<-time.After(100 * time.Millisecond)
		wg.Done()
	}()

	go func() {
		lock := New("bird", "owl", "snake")
		lock.Lock()
		defer lock.Unlock()

		<-time.After(100 * time.Millisecond)
		wg.Done()
	}()

	go func() {
		lock := New("owl", "snake")
		lock.Lock()
		defer lock.Unlock()

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
		lock := New("A", "C")
		lock.Lock()
		defer lock.Unlock()

		for resources["ac"] == 0 {
			lock.Yield()
		}
		resources["dc"] = 10

		wg.Done()
	}()

	go func() {
		lock := New("D", "C")
		lock.Lock()
		defer lock.Unlock()

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
				lock := New("A", "B", "C", "E", "Z")
				lock.Lock()
				<-time.After(30 * time.Millisecond)
				lock.Unlock()
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
				lock := New("B", "C", "K", "L", "Z")
				lock.Lock()
				<-time.After(200 * time.Millisecond)
				lock.Unlock()
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

func TestBankAccountProblem(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(3)

	joe := 50.0
	susan := 100.0

	// withdraw $80 from joe, only if balance is sufficient
	go func() {
		lock := New("joe")
		lock.Lock()
		defer lock.Unlock()

		for joe < 80.0 {
			lock.Yield()
		}
		joe -= 80.0

		wg.Done()
	}()

	// transfer $200 from susan to joe, only if balance is sufficient
	go func() {
		lock := New("joe", "susan")
		lock.Lock()
		defer lock.Unlock()

		for susan < 200.0 {
			lock.Yield()
		}

		susan -= 200.0
		joe += 200.0

		wg.Done()
	}()

	// susan deposit $300 to cover balance
	go func() {
		lock := New("susan")
		lock.Lock()
		defer lock.Unlock()

		susan += 300.0

		wg.Done()
	}()

	wg.Wait()
	assert.Equal(t, 170.0, joe)
	assert.Equal(t, 200.0, susan)
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
