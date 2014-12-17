package squaredance

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func stoporfail(t *testing.T, l Caller) error {
	donechan := make(chan error)
	go func() {
		l.Stop()
		donechan <- l.Wait()
		close(donechan)
	}()

	var err error
	select {
	case err = <-donechan:
	case <-time.After(time.Second):
		t.Fatal("Error stopping child")
	}
	return err
}

func TestSimpleFollow(t *testing.T) {
	l := NewCaller()
	sendchan := make(chan int)
	stuff := []int{}
	l.Spawn(func(f Follower) (err error) {
		for {
			select {
			case <-f.StopChan():
				return
			case v := <-sendchan:
				stuff = append(stuff, v)
			}
		}
		return
	})

	for _, v := range []int{1, 2, 3} {
		select {
		case sendchan <- v:
		case <-time.After(time.Second):
			t.Fatal("Send chan not accepting values")
		}
	}

	stoporfail(t, l)
	if len(stuff) != 3 {
		t.Fatalf("Expected three items, got %d", len(stuff))
	}
}

func TestMultipleWorkersNoErrors(t *testing.T) {
	l := NewCaller()
	tmu := sync.Mutex{}
	thing := 0
	for i := 0; i < 5; i++ {
		l.Spawn(func(f Follower) error {
			<-f.StopChan()
			tmu.Lock()
			defer tmu.Unlock()
			thing++
			return nil
		})
	}

	stoporfail(t, l)

	if thing != 5 {
		t.Fatalf("Expected 5, got %d", thing)
	}
}

func TestChildFailure(t *testing.T) {
	l := NewCaller()
	e := fmt.Errorf("Whoa. Error.")
	for i := 0; i < 5; i++ {
		l.Spawn(func(f Follower) error {
			<-f.StopChan()
			return nil
		})
	}

	l.Spawn(func(f Follower) error {
		return e
	})

	err := stoporfail(t, l)
	if err != e {
		t.Fatalf("Unexpected error: %v", e)
	}
}

func TestWaitBeforeStop(t *testing.T) {
	l := NewCaller()
	e := fmt.Errorf("Whoa. Error.")
	l.Spawn(func(f Follower) error {
		// Blocks until stop is called
		<-f.StopChan()
		return e
	})

	ec := make(chan error)
	go func() {
		ec <- l.Wait()
	}()

	l.Stop()
	select {
	case err := <-ec:
		if err != e {
			t.Fatalf("Unexpected error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for child to close")
	}
}

// Task spawns another task; both return immediately
func TestNested(t *testing.T) {
	l := NewCaller()
	e := fmt.Errorf("Whoa. Error")
	l.Spawn(func(f Follower) error {
		f.Spawn(func(Follower) error {
			return e
		})
		return nil
	})

	ec := make(chan error)
	go func() {
		ec <- l.Wait()
	}()
	var err error
	select {
	case err = <-ec:
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for task shutdown")
	}

	if err != e {
		t.Fatalf("Unexpected error (expected %v): %v", e, err)
	}
}

// Task spawns another task; parent blocks, child returns error
func TestNestedWaitErr(t *testing.T) {
	l := NewCaller()
	e := fmt.Errorf("Whoa. Error")

	l.Spawn(func(f Follower) error {
		f.Spawn(func(Follower) error {
			return e
		})
		<-f.StopChan()
		return nil
	})

	ec := make(chan error)
	go func() {
		ec <- l.Wait()
	}()
	var err error
	select {
	case err = <-ec:
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for task shutdown")
	}

	if err != e {
		t.Fatalf("Unexpected error (expected %v): %v", e, err)
	}
}
