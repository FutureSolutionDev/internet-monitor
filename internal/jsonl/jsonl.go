package jsonl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
)

const maxLine = 1 << 20

func Scan(path string, fn func(line []byte) bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return scanReader(f, fn)
}

func scanReader(r io.Reader, fn func(line []byte) bool) error {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), maxLine)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 || !json.Valid(line) {
			continue
		}
		if !fn(line) {
			return nil
		}
	}
	if err := sc.Err(); err != nil && !errors.Is(err, bufio.ErrTooLong) {
		return err
	}
	return nil
}

func Tail(path string, limit int) ([]json.RawMessage, error) {
	if limit <= 0 {
		return []json.RawMessage{}, nil
	}
	ring := make([]json.RawMessage, 0, limit)
	start := 0
	err := Scan(path, func(line []byte) bool {
		cp := append(json.RawMessage(nil), line...)
		if len(ring) < limit {
			ring = append(ring, cp)
		} else {
			ring[start] = cp
			start = (start + 1) % limit
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	out := make([]json.RawMessage, 0, len(ring))
	for i := len(ring) - 1; i >= 0; i-- {
		out = append(out, ring[(start+i)%len(ring)])
	}
	return out, nil
}

func Decode[T any](path string, fn func(T)) error {
	return Scan(path, func(line []byte) bool {
		var v T
		if json.Unmarshal(line, &v) == nil {
			fn(v)
		}
		return true
	})
}
