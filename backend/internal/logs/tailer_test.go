package logs

import (
	"reflect"
	"testing"
)

func TestNewLinesReturnsOnlyAppendedLines(t *testing.T) {
	previous := []string{"one", "two"}
	current := []string{"one", "two", "three", "four"}

	got := NewLines(previous, current)
	want := []string{"three", "four"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NewLines() = %v, want %v", got, want)
	}
}

func TestNewLinesKeepsFreshLinesAfterLogRotation(t *testing.T) {
	previous := []string{"old one", "old two"}
	current := []string{"new one", "new two"}

	got := NewLines(previous, current)
	if !reflect.DeepEqual(got, current) {
		t.Fatalf("NewLines() = %v, want %v", got, current)
	}
}
