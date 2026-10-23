package jsonl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLines(t *testing.T, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "x.jsonl")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTailReturnsNewestFirstAndSkipsPartialLines(t *testing.T) {
	path := writeLines(t, `{"n":1}`, `{"n":2}`, `garbage`, `{"n":3}`, `{"n":4}`, `{"n":5`)
	got, err := Tail(path, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{`{"n":4}`, `{"n":3}`, `{"n":2}`}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %s", len(got), len(want), got)
	}
	for i := range want {
		if string(got[i]) != want[i] {
			t.Errorf("got[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestTailFewerLinesThanLimit(t *testing.T) {
	got, err := Tail(writeLines(t, `{"n":1}`, `{"n":2}`), 10)
	if err != nil || len(got) != 2 || string(got[0]) != `{"n":2}` {
		t.Fatalf("got %s err=%v", got, err)
	}
}

func TestTailMissingFile(t *testing.T) {
	if _, err := Tail(filepath.Join(t.TempDir(), "missing.jsonl"), 5); !os.IsNotExist(err) {
		t.Errorf("err = %v, want not-exist", err)
	}
}

func TestDecode(t *testing.T) {
	var sum int
	err := Decode(writeLines(t, `{"n":1}`, `{"n":2}`, `{"n":"bad"}`), func(v struct{ N int }) { sum += v.N })
	if err != nil || sum != 3 {
		t.Fatalf("sum = %d err=%v", sum, err)
	}
}

func TestScanSkipsOversizedLineAndContinues(t *testing.T) {
	huge := `{"x":"` + strings.Repeat("a", maxLine+10) + `"}`
	got, err := Tail(writeLines(t, `{"n":1}`, huge, `{"n":2}`), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || string(got[0]) != `{"n":2}` {
		t.Fatalf("got %d entries, first=%s", len(got), got[0])
	}
}
