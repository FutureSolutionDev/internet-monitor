package alerts

import (
	"encoding/json"
	"internet-monitor/internal/jsonl"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	HistoryPrefix   = "notifications_"
	historyState    = "notifications_state.json"
	historyCapacity = 500
	MaxListLimit    = historyCapacity
)

type History struct {
	mu       sync.Mutex
	dir      func() string
	items    []Record
	lastRead time.Time
	logf     func(format string, args ...interface{})
}

type historyFile struct {
	LastReadAt time.Time `json:"last_read_at"`
}

func OpenHistory(dir func() string, logf func(format string, args ...interface{})) *History {
	h := &History{dir: dir, logf: logf}
	h.load()
	return h
}

func (h *History) log(format string, args ...interface{}) {
	if h.logf != nil {
		h.logf(format, args...)
	}
}

func (h *History) load() {
	dir := h.dir()
	if data, err := os.ReadFile(filepath.Join(dir, historyState)); err == nil {
		var st historyFile
		if json.Unmarshal(data, &st) == nil {
			h.lastRead = st.LastReadAt
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var files []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasPrefix(name, HistoryPrefix) && strings.HasSuffix(name, ".jsonl") {
			files = append(files, name)
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	var chunks [][]Record
	total := 0
	for _, name := range files {
		var day []Record
		_ = jsonl.Scan(filepath.Join(dir, name), func(line []byte) bool {
			var r Record
			if json.Unmarshal(line, &r) == nil && r.ID != "" {
				day = append(day, r)
			}
			return true
		})
		chunks = append(chunks, day)
		total += len(day)
		if total >= historyCapacity {
			break
		}
	}
	for i := len(chunks) - 1; i >= 0; i-- {
		h.items = append(h.items, chunks[i]...)
	}
	if len(h.items) > historyCapacity {
		h.items = h.items[len(h.items)-historyCapacity:]
	}
}

func (h *History) Add(r Record) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.items = append(h.items, r)
	if len(h.items) > historyCapacity {
		h.items = append([]Record(nil), h.items[len(h.items)-historyCapacity:]...)
	}
	data, err := json.Marshal(r)
	if err != nil {
		return
	}
	path := filepath.Join(h.dir(), HistoryPrefix+r.Time.Format("2006-01-02")+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		h.log("ERROR open notification history: %v", err)
		return
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		h.log("ERROR write notification history: %v", err)
	}
}

func (h *History) List(limit int) []Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	if limit <= 0 || limit > len(h.items) {
		limit = len(h.items)
	}
	out := make([]Record, 0, limit)
	for i := len(h.items) - 1; i >= 0 && len(out) < limit; i-- {
		out = append(out, h.items[i])
	}
	return out
}

func (h *History) Unread() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.unreadLocked()
}

func (h *History) unreadLocked() int {
	n := 0
	for i := len(h.items) - 1; i >= 0; i-- {
		if !h.items[i].Time.After(h.lastRead) {
			break
		}
		n++
	}
	return n
}

func (h *History) LastRead() time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.lastRead
}

func (h *History) MarkRead(at time.Time) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	if at.After(h.lastRead) {
		h.lastRead = at
		data, _ := json.Marshal(historyFile{LastReadAt: at})
		path := filepath.Join(h.dir(), historyState)
		tmp := path + ".tmp"
		err := os.WriteFile(tmp, data, 0o644)
		if err == nil {
			err = os.Rename(tmp, path)
		}
		if err != nil {
			h.log("ERROR save notification read state: %v", err)
		}
	}
	return h.unreadLocked()
}
