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
	br := bufio.NewReaderSize(r, 64*1024)
	var buf []byte
	skipping := false
	for {
		chunk, err := br.ReadSlice('\n')
		if !skipping {
			if len(buf)+len(chunk) > maxLine {
				skipping, buf = true, buf[:0]
			} else {
				buf = append(buf, chunk...)
			}
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		if !skipping {
			line := bytes.TrimSpace(buf)
			if len(line) > 0 && json.Valid(line) && !fn(line) {
				return nil
			}
		}
		buf, skipping = buf[:0], false
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
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
