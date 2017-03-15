# Multilock

[![Build Status](https://travis-ci.org/atedja/go-multilock.svg?branch=master)](https://travis-ci.org/atedja/go-multilock)

Multilock allows you to obtain multiple locks without deadlock. It also uses
strings as locks, which allows multiple goroutines to synchronize independently
without having to share common mutex objects.

One common application is to use external ids (e.g. resource ids, filenames,
database ids) as the lock, and thereby preventing multiple goroutines from 
potentially reading/writing to the same resources, creating some form of 
transactional locking.

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

  john := 1000
  susan := 2000

  go func() {
    lock := multilock.New("john", "susan")
    lock.Lock()
    defer lock.Unlock()

    fmt.Println("Transferring money from john to susan")
    john -= 200
    susan += 200
    wg.Done()
  }()

  go func() {
    lock := multilock.New("susan", "john")
    lock.Lock()
    defer lock.Unlock()

    fmt.Println("Transferring money from susan to john")
    john += 400
    susan -= 400
    wg.Done()
  }()

  fmt.Println("john's balance", john)
  fmt.Println("susan's balance", susan)

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
again, Multilock will automatically create new locks for those keys and 
everybody is happy again.

    multilock.Clean()

#### Compatibility with `sync.Locker` interface

`multilock.Lock` implements `sync.Locker` interface, and can be used by other 
locking mechanism, e.g. `sync.Cond`.

### Best Practices

#### Specify all your locks at once

Specify all the locks you need for your transaction at once. DO NOT create
nested `Lock()` statements.  The lock is not reentrant. Likewise, you should
not call `Unlock()` without a matching `Lock()`. They are both blocking.

#### Always `Unlock` your locks

Just common sense.
