# Multilock

[![Build Status](https://travis-ci.org/atedja/go-multilock.svg?branch=master)](https://travis-ci.org/atedja/go-multilock)

Multilock allows you to obtain multiple locks without deadlock. It also uses
strings as locks, which allows multiple goroutines to synchronize independently
without having to share common mutex objects.

One common application is to use multiple external ids (e.g. resource IDs)
as the lock, and thereby preventing multiple goroutines from potentially
reading/writing to the same resources, creating some form of transactional
locking.

### Installation

    go get github.com/atedja/go-multilock

### Example

```go
package main

import (
  "fmt"
  "sync"
  "github.com/atedja/go-multilock"
)

func main() {
  var wg sync.WaitGroup
  wg.Add(2)

  go func() {
    lock := multilock.New("bird", "dog", "cat")
    lock.Lock()
    defer lock.Unlock()

    fmt.Println("Taking photos of bird, dog, and cat...")
    <-time.After(1 * time.Second)
    fmt.Println("bird, dog, and cat photos taken!")
    wg.Done()
  }()

  go func() {
    lock := multilock.New("whale", "cat")
    lock.Lock()
    defer lock.Unlock()

    fmt.Println("Taking photos of whale and cat...")
    <-time.After(1 * time.Second)
    fmt.Println("whale and cat photos taken!")
    wg.Done()
  }()

  wg.Wait()
}
```

#### [Full API Documentation](https://godoc.org/github.com/atedja/go-multilock)

### Basic Usage

#### Lock and Unlock

    lock := multilock.New("somekey")
    lock.Lock()
    defer lock.Unlock()

#### Yield

Temporarily unlocks the acquired lock, yields cpu time to other goroutines,
then attempts to lock the same keys again.

    lock := multilock.New("somekey")
    lock.Lock()
    for resource["somekey"] == nil {
      lock.Yield()
    }
    // process resource["somekey"]
    Unlock(lock)

#### Clean your unused locks

If you use nondeterministic number of keys, e.g. timestamp, then overtime the
number of locks will grow creating a memory "leak". `Clean()` will remove
unused locks. This method is threadsafe, and can be executed even while other
goroutines furiously attempt to acquire other keys. If some keys are to be used
again (as soon as immediately), Multilock will automatically create new locks
for those keys and everybody is happy again.

    multilock.Clean()

#### Compatibility with `sync.Locker` interface

For compatibilty's sake, `multilock.Lock` implements `sync.Locker` interface,
and can be used by other locking mechanism, e.g. `sync.Cond`.

### Best Practices

#### Specify all your locks at once

Specify all the locks you need for your transaction at once. DO NOT create
nested `Lock()` statements.  The lock is NOT reentrant. Likewise, DO NOT
call `Unlock()` without a matching `Lock()`. They both are blocking.

#### Always `Unlock` your locks

Just common sense.
