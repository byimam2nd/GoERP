package event

import (
	"sync"
)

type Event struct {
	Name string
	Data interface{}
}

type Handler func(e Event)

type EventBus struct {
	handlers map[string][]Handler
	mu       sync.RWMutex
}

var DefaultBus = &EventBus{
	handlers: make(map[string][]Handler),
}

func (b *EventBus) Subscribe(eventName string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], h)
}

func (b *EventBus) Publish(eventName string, data interface{}) {
	b.mu.RLock()
	handlers, exists := b.handlers[eventName]
	b.mu.RUnlock()

	if !exists {
		return
	}

	e := Event{Name: eventName, Data: data}
	for _, h := range handlers {
		// Run in goroutine for async behavior
		go h(e)
	}
}
