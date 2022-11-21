package models

import (
	"sync"
	"time"
)

type Event struct {
	When       time.Time
	NotifiedAt time.Time
}

type CustomerEvent struct {
	Email      string
	Message    string
	Closed     bool
	Events     []*Event
	processing bool
	sync.Mutex
}

func (c *CustomerEvent) Done() {
	c.Closed = true
}

func (c *CustomerEvent) Start() {
	c.Lock()
	c.processing = true
	c.Unlock()
}
func (c *CustomerEvent) Stop() {
	c.Lock()
	c.processing = false
	c.Unlock()
}

func (c *CustomerEvent) IsProcessing() bool {
	return c.processing
}
