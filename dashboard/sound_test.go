package dashboard

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func soundUpload(t *testing.T, s *Server, payload []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("sound", "upload.bin")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write(payload)
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/notification-sound", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	s.serveNotificationSound(rec, req)
	return rec
}

func soundIsCustom(s *Server) bool {
	rec := httptest.NewRecorder()
	s.serveNotificationSound(rec, httptest.NewRequest(http.MethodHead, "/notification-sound", nil))
	return rec.Header().Get("X-Custom-Sound") == "1"
}

func TestSoundUploadRejectsUnsupportedWithoutFakeSuccess(t *testing.T) {
	dir := t.TempDir()
	s := NewServer(0, filepath.Join(dir, "config.json"), dir, "test", nil)

	rec := soundUpload(t, s, bytes.Repeat([]byte("OggS"), 64))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body=%s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "notification.wav")); err == nil {
		t.Error("custom WAV must not exist after a rejected upload")
	}
	if soundIsCustom(s) {
		t.Error("HEAD reports a custom sound after a rejected upload")
	}
}

func TestSoundUploadMP3BecomesCustomWAV(t *testing.T) {
	dir := t.TempDir()
	s := NewServer(0, filepath.Join(dir, "config.json"), dir, "test", nil)

	rec := soundUpload(t, s, RingtoneMp3())
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	wav, err := os.ReadFile(filepath.Join(dir, "notification.wav"))
	if err != nil || string(wav[0:4]) != "RIFF" {
		t.Fatalf("custom WAV not written: %v", err)
	}
	if !soundIsCustom(s) {
		t.Error("HEAD does not report the uploaded custom sound")
	}

	del := httptest.NewRecorder()
	s.serveNotificationSound(del, httptest.NewRequest(http.MethodDelete, "/notification-sound", nil))
	if soundIsCustom(s) {
		t.Error("custom sound still reported after reset")
	}
}

func TestSoundUploadAcceptsWAV(t *testing.T) {
	dir := t.TempDir()
	s := NewServer(0, filepath.Join(dir, "config.json"), dir, "test", nil)

	rec := soundUpload(t, s, NotificationWav())
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	got, _ := os.ReadFile(filepath.Join(dir, "notification.wav"))
	if !bytes.Equal(got, NotificationWav()) {
		t.Error("uploaded WAV was not stored verbatim")
	}
}
