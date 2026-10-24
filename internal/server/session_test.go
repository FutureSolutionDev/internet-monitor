package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSessionKeySurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	first := loadOrCreateToken(dir, now)
	if !wellFormedToken(first) {
		t.Fatalf("token %q is not 64 hex chars", first)
	}
	if again := loadOrCreateToken(dir, now.Add(time.Hour)); again != first {
		t.Fatal("a restart within a day changed the key")
	}
	info, err := os.Stat(filepath.Join(dir, sessionKey))
	if err != nil {
		t.Fatal(err)
	}
	if later := loadOrCreateToken(dir, info.ModTime().Add(sessionKeyTTL+time.Minute)); later == first {
		t.Fatal("a key older than a day was reused")
	}
}

func TestSessionKeyRejectsTamperedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, sessionKey)
	for _, bad := range []string{"", "short", "zz" + string(make([]byte, 62))} {
		if err := os.WriteFile(path, []byte(bad), 0o600); err != nil {
			t.Fatal(err)
		}
		if tok := loadOrCreateToken(dir, time.Now()); tok == bad || !wellFormedToken(tok) {
			t.Fatalf("bad key %q was used", bad)
		}
	}
}

func TestSessionKeyWithoutAppDir(t *testing.T) {
	if a, b := loadOrCreateToken("", time.Now()), loadOrCreateToken("", time.Now()); a == b || !wellFormedToken(a) {
		t.Fatal("tokens without an app dir must be fresh and well formed")
	}
}
