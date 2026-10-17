package chat

import "testing"

func TestTellrawCommandEscapesUntrustedMessage(t *testing.T) {
	command, err := TellrawCommand("hello \"friend\" \\ newline\n")
	if err != nil { t.Fatal(err) }
	want := `tellraw @a {"text":"[Dashboard Admin] hello \"friend\" \\ newline"}`
	if command != want { t.Fatalf("got %q, want %q", command, want) }
}

func TestTellrawCommandRejectsControlCharactersAndLongMessages(t *testing.T) {
	if _, err := TellrawCommand("bad\x00message"); err == nil { t.Fatal("expected NUL character rejection") }
	if _, err := TellrawCommand(string(make([]byte, 257))); err == nil { t.Fatal("expected length rejection") }
}
