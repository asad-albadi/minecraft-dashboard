package backups

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListReadOnlyArchives(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "world.zip"), []byte("archive"), 0600)
	os.WriteFile(filepath.Join(d, "secret.env"), []byte("secret"), 0600)
	os.Symlink("world.zip", filepath.Join(d, "link.zip"))
	got, err := List(d)
	if err != nil || len(got) != 1 || got[0].Name != "world.zip" || got[0].Size != 7 || got[0].Integrity != "Unknown" || got[0].ModifiedAt.IsZero() {
		t.Fatalf("%+v %v", got, err)
	}
}
