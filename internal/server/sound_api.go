package server

import (
	"internet-monitor/internal/assets"
	"internet-monitor/internal/audio"
	"internet-monitor/internal/sound"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const maxSoundUpload = 10 << 20

func (s *Server) customWavPath() string {
	return filepath.Join(s.appDir, sound.CustomFileName)
}

func (s *Server) legacyMp3Path() string {
	return filepath.Join(s.appDir, "notification.mp3")
}

func (s *Server) serveSoundGet(w http.ResponseWriter, r *http.Request) {
	if data, err := os.ReadFile(s.customWavPath()); err == nil {
		w.Header().Set("X-Custom-Sound", "1")
		w.Header().Set("X-Sound-Duration", strconv.FormatFloat(audio.WAVDuration(data), 'f', 2, 64))
		w.Header().Set("Content-Type", "audio/wav")
		http.ServeFile(w, r, s.customWavPath())
		return
	}
	w.Header().Set("Content-Type", "audio/mpeg")
	if r.Method == http.MethodGet {
		_, _ = w.Write(assets.RingtoneMP3())
	}
}

func (s *Server) serveSoundUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxSoundUpload+4096)
	file, _, err := r.FormFile("sound")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_file", "invalid file")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxSoundUpload+1))
	if err != nil || len(data) == 0 || len(data) > maxSoundUpload {
		writeError(w, http.StatusBadRequest, "read_failed", "file is empty or larger than 10 MB")
		return
	}
	wav, err := audio.ToWAV(data)
	if err != nil {
		s.logf("custom sound: rejected upload: %v", err)
		writeError(w, http.StatusUnprocessableEntity, "unsupported", err.Error())
		return
	}
	if err := os.WriteFile(s.customWavPath(), wav, 0o644); err != nil {
		s.logf("custom sound: WAV save failed: %v", err)
		writeError(w, http.StatusInternalServerError, "save_failed", "save failed")
		return
	}
	_ = os.Remove(s.legacyMp3Path())
	s.broadcast("sound")
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) serveSoundReset(w http.ResponseWriter, _ *http.Request) {
	_ = os.Remove(s.legacyMp3Path())
	_ = os.Remove(s.customWavPath())
	s.broadcast("sound")
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
