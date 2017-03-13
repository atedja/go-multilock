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

		<-time.After(1 * time.Second)
		wg.Done()
	}()

	go func() {
		lock := Lock("cat", "dog", "bird")
		defer Unlock(lock)

		<-time.After(1 * time.Second)
		wg.Done()
	}()

	go func() {
		lock := Lock("cat", "bird", "owl")
		defer Unlock(lock)

		<-time.After(1 * time.Second)
		wg.Done()
	}()

	go func() {
		lock := Lock("bird", "owl", "snake")
		defer Unlock(lock)

		<-time.After(1 * time.Second)
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
