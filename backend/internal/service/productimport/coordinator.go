package productimport

import "sync"

type Coordinator struct {
	mu     sync.Mutex
	active map[int64]struct{}
}

func NewCoordinator() *Coordinator {
	return &Coordinator{
		active: make(map[int64]struct{}),
	}
}

func (c *Coordinator) Acquire(accountID int64) bool {
	if c == nil || accountID <= 0 {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.active[accountID]; exists {
		return false
	}

	c.active[accountID] = struct{}{}
	return true
}

func (c *Coordinator) Release(accountID int64) {
	if c == nil || accountID <= 0 {
		return
	}

	c.mu.Lock()
	delete(c.active, accountID)
	c.mu.Unlock()
}
