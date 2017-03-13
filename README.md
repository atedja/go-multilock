# Multilock

[![Build Status](https://travis-ci.org/atedja/go-multilock.svg?branch=master)](https://travis-ci.org/atedja/go-multilock)

Multilock allows you to obtain multiple locks without deadlock. It also uses
strings as locks, which allows multiple goroutines to synchronize independently
without having to share common mutex objects.

One common application is to use an external id (e.g. IDs from a database)
as the lock, and thereby preventing multiple goroutines from potentially
reading/writing to the same rows in the database.

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
    lock := multilock.Lock("bird", "dog", "cat")
    defer multilock.Unlock(lock)

    fmt.Println("Taking photos of bird, dog, and cat...")
    <-time.After(1 * time.Second)
    fmt.Println("bird, dog, and cat photos taken!")
    wg.Done()
  }()

  go func() {
    lock := multilock.Lock("whale", "cat")
    defer multilock.Unlock(lock)

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

    lock := multilock.Lock("somekey")
    defer multilock.Unlock(lock)

#### Yield

Temporarily unlocks the acquired lock, yields cpu time to other goroutines, then attempts to lock the same keys again.

    lock := multilock.Lock("somekey")
    for resource["somekey"] == nil {
      Yield(lock)
    }
    // process resource["somekey"]
    Unlock(lock)

#### Clean your unused locks

If you use undetermined number of keys, e.g. timestamp, then overtime the number of locks will grow creating a
memory "leak". `Clean()` will remove unused locks. This method is threadsafe. If some keys are to be used again,
the system will automatically create new locks for those keys and everybody is happy again.

    multilock.Clean()

### Best Practices

#### Specify all your locks at once

Specify all the locks you need at once. DO NOT create nested `multilock.Lock()`
statements.  It beats the purpose of having this library in the first place.

#### Always `Unlock` your locks

Just common sense.
