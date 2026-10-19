package httpapi

import "testing"

func TestEventBrokerCapsConnectionsAndBroadcasts(t *testing.T) {
	broker := newEventBroker(1)
	subscriber, cancel, ok := broker.subscribe()
	if !ok {
		t.Fatal("first subscriber was rejected")
	}
	defer cancel()
	if _, _, ok := broker.subscribe(); ok {
		t.Fatal("subscriber beyond the configured limit was accepted")
	}

	broker.publish(dashboardEvent{name: "console.line", payload: "message"})
	event := <-subscriber
	if event.name != "console.line" || event.payload != "message" {
		t.Fatalf("event = %#v, want a published console line", event)
	}
}
