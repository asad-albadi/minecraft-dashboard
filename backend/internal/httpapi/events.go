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
	reader := logs.NewReader(a.Config.LogPath, 500)
	defer reader.Close()
	initial := true
	for tick := 1; ; tick++ {
		select {
		case <-a.stop:
			return
		case <-ticker.C:
		}
		if a.Config.LogPath == "" {
			continue
		}
		current, err := reader.Next()
		a.logMu.Lock()
		a.logError = err != nil
		a.logMu.Unlock()
		if err == nil {
			a.recordLines(current)
			var e error
			if initial {
				e = a.History.Reconcile(current)
			} else {
				e = a.History.Append(current)
			}
			if e != nil {
				a.logMu.Lock()
				a.historyError = true
				a.logMu.Unlock()
			}
			initial = false
		}
		for _, line := range current {
			a.Events.publish(dashboardEvent{name: "console.line", payload: map[string]string{"line": line}})
		}
		if len(current) > 0 {
			a.Events.publish(dashboardEvent{name: "console.snapshot", payload: a.consoleSnapshot()})
		}
		if tick%15 == 0 && a.Events.hasSubscribers() {
			a.Events.publish(dashboardEvent{name: "server.status", payload: a.statusPayload(context.Background())})
		}
	}
}

type consoleSnapshot struct {
	Sequence uint64   `json:"sequence"`
	Lines    []string `json:"lines"`
}

func (a *API) recordLines(lines []string) {
	if len(lines) == 0 {
		return
	}
	a.logMu.Lock()
	defer a.logMu.Unlock()
	a.lines = append(a.lines, lines...)
	if len(a.lines) > 500 {
		a.lines = a.lines[len(a.lines)-500:]
	}
	a.logSequence++
}
func (a *API) consoleSnapshot() consoleSnapshot {
	a.logMu.RLock()
	defer a.logMu.RUnlock()
	return consoleSnapshot{a.logSequence, append([]string{}, a.lines...)}
}
