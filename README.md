# Squaredance

Simple task coordination

Inspired by [Context](https://blog.golang.org/context), but with specific error handling
semantics

[![Build Status](https://travis-ci.org/lytics/squaredance.svg?branch=master)](https://travis-ci.org/lytics/squaredance) [![GoDoc](https://godoc.org/github.com/lytics/squaredance?status.svg)](https://godoc.org/github.com/lytics/squaredance)

### Simple case
```go
c := squaredance.NewCaller()

vals := make(chan string)
// Start 5 tasks
for i := 0; i < 5; i++ {
	c.Spawn(func(sf squaredance.Follower) (err error) {
		for {
			select {
			case <-sf.StopChan():
				return
			case val := <-vals:
				fmt.Printf("%s\n", val)
			}
		}
	})

}

for _, v := range []string{"swing", "your", "partner", "round", "and", "round"} {
	vals <- v
}

c.Stop()
c.Wait()
```

### Nested peer and return error
```go
c := squaredance.NewCaller()

c.Spawn(func(sf squaredance.Follower) error {
	val := make(chan string)
	sf.Spawn(func(sf squaredance.Follower) error {
		// This task returns immediately
		v := <-val
		fmt.Printf("Got a thing: %s\n", v)
		return fmt.Errorf("Fell down")
	})
	val <- "Spin"
	// This task waits until stop is called
	<-sf.StopChan()
	return nil
})

c.Stop()
err := c.Wait()
fmt.Printf("%v\n", err)
```
