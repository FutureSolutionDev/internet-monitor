package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrConflict = errors.New("configuration was changed elsewhere; reload and try again")

type Store struct {
	mu      sync.RWMutex
	path    string
	cfg     Config
	rev     uint64
	loadErr error
	subs    []func(Config)
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, cfg: Default(), rev: 1}
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return s, s.write(s.cfg)
	case err != nil:
		return s, err
	}
	cfg, perr := Parse(data)
	if perr != nil {
		backup := fmt.Sprintf("%s.invalid-%s", path, time.Now().Format("20060102-150405"))
		_ = os.WriteFile(backup, data, 0o600)
		s.loadErr = fmt.Errorf("config.json could not be parsed (%v); a copy was saved to %s and defaults are in use", perr, filepath.Base(backup))
		return s, nil
	}
	s.cfg = cfg
	return s, nil
}

func (s *Store) Path() string { return s.path }

func (s *Store) LoadError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadErr
}

func (s *Store) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Clone()
}

func (s *Store) Snapshot() (Config, uint64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Clone(), s.rev
}

func (s *Store) Revision() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rev
}

func (s *Store) Subscribe(fn func(Config)) {
	s.mu.Lock()
	s.subs = append(s.subs, fn)
	s.mu.Unlock()
}

func (s *Store) Update(expectRev uint64, fn func(*Config) error) (Config, uint64, error) {
	s.mu.Lock()
	if expectRev != 0 && expectRev != s.rev {
		cur, rev := s.cfg.Clone(), s.rev
		s.mu.Unlock()
		return cur, rev, ErrConflict
	}
	next := s.cfg.Clone()
	if err := fn(&next); err != nil {
		s.mu.Unlock()
		return Config{}, 0, err
	}
	next.Sanitize()
	if err := next.Validate(); err != nil {
		s.mu.Unlock()
		return Config{}, 0, err
	}
	if err := s.write(next); err != nil {
		s.mu.Unlock()
		return Config{}, 0, err
	}
	s.cfg = next
	s.rev++
	s.loadErr = nil
	rev := s.rev
	subs := append([]func(Config){}, s.subs...)
	s.mu.Unlock()

	for _, sub := range subs {
		sub(next.Clone())
	}
	return next.Clone(), rev, nil
}

func (s *Store) write(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
