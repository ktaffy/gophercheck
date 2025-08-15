package watcher

import (
	"fmt"
	"sync"
	"time"
)

type debouncer struct {
	delay    time.Duration
	events   map[string]FileChangeEvent
	timer    *time.Timer
	mutex    sync.Mutex
	stopChan chan struct{}
}

func newDebouncer(delay time.Duration) *debouncer {
	return &debouncer{
		delay:    delay,
		events:   make(map[string]FileChangeEvent),
		stopChan: make(chan struct{}),
	}
}

func (d *debouncer) add(event FileChangeEvent, handler FileChangeHandler) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.events[event.Path] = event
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.delay, func() {
		d.flush(handler)
	})
}

func (d *debouncer) flush(handler FileChangeHandler) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if len(d.events) == 0 {
		return
	}
	changedFiles := make([]string, 0, len(d.events))
	for path := range d.events {
		changedFiles = append(changedFiles, path)
	}
	d.events = make(map[string]FileChangeEvent)
	if err := handler(changedFiles); err != nil {
		// Will add better error handling later on for now just print
		fmt.Printf("Handler error: %v\n", err)
	}
}

func (d *debouncer) stop() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.timer != nil {
		d.timer.Stop()
	}
	close(d.stopChan)
}
