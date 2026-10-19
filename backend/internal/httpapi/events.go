package httpapi

import (
	"context"
	"minecraft-dashboard/backend/internal/logs"
	"sync"
	"time"
)

type dashboardEvent struct {
	name    string
	payload any
}

type eventBroker struct {
	mu          sync.RWMutex
	subscribers map[chan dashboardEvent]struct{}
	maximum     int
}

func newEventBroker(maximum int) *eventBroker {
	return &eventBroker{subscribers: make(map[chan dashboardEvent]struct{}), maximum: maximum}
}

func (b *eventBroker) subscribe() (chan dashboardEvent, func(), bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.subscribers) >= b.maximum {
		return nil, nil, false
	}
	subscriber := make(chan dashboardEvent, 32)
	b.subscribers[subscriber] = struct{}{}
	var once sync.Once
	return subscriber, func() { once.Do(func() { b.unsubscribe(subscriber) }) }, true
}

func (b *eventBroker) unsubscribe(subscriber chan dashboardEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subscribers[subscriber]; ok {
		delete(b.subscribers, subscriber)
		close(subscriber)
	}
}

func (b *eventBroker) hasSubscribers() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers) > 0
}

func (b *eventBroker) publish(event dashboardEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for subscriber := range b.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}

func (a *API) pollEvents() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	previous := a.logTail()
	for tick := 1; ; tick++ {
		<-ticker.C
		if !a.Events.hasSubscribers() {
			previous = a.logTail()
			continue
		}
		current := a.logTail()
		for _, line := range logs.NewLines(previous, current) {
			a.Events.publish(dashboardEvent{name: "console.line", payload: map[string]string{"line": line}})
		}
		previous = current
		if tick%15 == 0 {
			a.Events.publish(dashboardEvent{name: "server.status", payload: a.statusPayload(context.Background())})
		}
	}
}
