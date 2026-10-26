package history

import "testing"

func TestReconcileRetainsMessagesWrittenDuringDowntime(t *testing.T) {
	s, _ := NewStore(t.TempDir())
	first := "[12:00:00] [Server thread/INFO]: <Alex> before"
	next := "[12:00:01] [Server thread/INFO]: <Alex> during downtime"
	if err := s.Append([]string{first}); err != nil {
		t.Fatal(err)
	}
	if err := s.Reconcile([]string{first, next}); err != nil {
		t.Fatal(err)
	}
	if len(s.Chat()) != 2 || s.Chat()[1].Message != "during downtime" {
		t.Fatal(s.Chat())
	}
}
