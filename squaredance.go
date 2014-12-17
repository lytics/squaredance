// Squaredance provides a simple interface for spawning stoppable child tasks
package squaredance

import "sync"

// The Caller allows spawning, stopping and waiting on child tasks
type Caller interface {
	// Wait blocks until all applicable children close
	Wait() error
	// Stops all children
	Stop()
	// Spawns a child task
	Spawn(func(Follower) error)
}

// The follower is passed to the spawned function, allowing access to
// the stop channel, as well as spawning additional peers
type Follower interface {
	// Closed when the following task should stop
	StopChan() <-chan struct{}
	// Spawn a new (peer) child task
	Spawn(func(Follower) error)
}

// Create a new parent Caller
func NewCaller() Caller {
	return &caller{
		stopchan: make(chan struct{}),
	}
}

type caller struct {
	childwg  sync.WaitGroup
	mu       sync.Mutex
	stopchan chan struct{}
	firsterr error
	stopped  bool
}

func (t *caller) Spawn(sf func(Follower) error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped {
		return
	}
	t.childwg.Add(1)
	go func() {
		err := sf(t)
		if err != nil {
			t.setError(err)
		}
		t.childwg.Done()
	}()
}

func (t *caller) Wait() error {
	t.childwg.Wait()
	return t.firsterr
}

func (t *caller) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.stopped {
		return
	}
	t.stop()
}

func (t *caller) StopChan() <-chan struct{} {
	return t.stopchan
}

func (t *caller) setError(e error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.firsterr == nil {
		t.firsterr = e
		t.stop()
	}
}

// Any calling methods must have the mutex acquired
func (t *caller) stop() {
	if !t.stopped {
		t.stopped = true
		close(t.stopchan)
	}
}
