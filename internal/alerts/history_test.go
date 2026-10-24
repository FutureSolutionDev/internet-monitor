package alerts

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistoryPersistsAndCountsUnread(t *testing.T) {
	dir := t.TempDir()
	h := OpenHistory(func() string { return dir }, nil)
	day1 := time.Date(2026, 10, 23, 23, 0, 0, 0, time.UTC)
	day2 := day1.Add(2 * time.Hour)
	h.Add(Record{ID: "a", Time: day1, Kind: KindDown, Sent: []Channel{ChannelDesktop}})
	h.Add(Record{ID: "b", Time: day2, Kind: KindRestored, Sent: []Channel{}})
	h.Add(Record{ID: "c", Time: day2.Add(time.Minute), Kind: KindTest, Sent: []Channel{}})

	if got := h.List(2); len(got) != 2 || got[0].ID != "c" || got[1].ID != "b" {
		t.Fatalf("list = %+v", got)
	}
	if h.Unread() != 3 {
		t.Fatalf("unread = %d", h.Unread())
	}
	if left := h.MarkRead(day2); left != 1 {
		t.Fatalf("unread after mark = %d", left)
	}
	if left := h.MarkRead(day1); left != 1 {
		t.Fatalf("marking an older time moved the cursor back: %d", left)
	}

	files, _ := filepath.Glob(filepath.Join(dir, HistoryPrefix+"*.jsonl"))
	if len(files) != 2 {
		t.Fatalf("history files = %v", files)
	}

	again := OpenHistory(func() string { return dir }, nil)
	if got := again.List(0); len(got) != 3 || got[0].ID != "c" || got[2].ID != "a" {
		t.Fatalf("reloaded = %+v", got)
	}
	if again.Unread() != 1 || !again.LastRead().Equal(day2) {
		t.Fatalf("reloaded unread=%d lastRead=%v", again.Unread(), again.LastRead())
	}
}

func TestHistoryKeepsTheNewestEntries(t *testing.T) {
	dir := t.TempDir()
	h := OpenHistory(func() string { return dir }, nil)
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < historyCapacity+20; i++ {
		h.Add(Record{ID: string(rune('a'+i%26)) + time.Duration(i).String(), Time: start.Add(time.Duration(i) * time.Hour), Kind: KindDown})
	}
	if got := len(h.List(0)); got != historyCapacity {
		t.Fatalf("kept %d", got)
	}
	again := OpenHistory(func() string { return dir }, nil)
	list := again.List(0)
	if len(list) != historyCapacity || !list[0].Time.Equal(start.Add(time.Duration(historyCapacity+19)*time.Hour)) {
		t.Fatalf("reloaded %d, newest %v", len(list), list[0].Time)
	}
}

func TestHistoryIgnoresBrokenLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, HistoryPrefix+"2026-10-24.jsonl")
	if err := os.WriteFile(path, []byte("not json\n{\"id\":\"x\",\"time\":\"2026-10-24T10:00:00Z\",\"kind\":\"down\"}\n{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := OpenHistory(func() string { return dir }, nil).List(0); len(got) != 1 || got[0].ID != "x" {
		t.Fatalf("list = %+v", got)
	}
}
