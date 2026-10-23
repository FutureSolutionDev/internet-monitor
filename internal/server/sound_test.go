package server

import (
	"bytes"
	"internet-monitor/internal/assets"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func soundUpload(t *testing.T, e *testEnv, payload []byte) *http.Response {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("sound", "upload.bin")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write(payload)
	_ = mw.Close()
	req, _ := http.NewRequest(http.MethodPost, e.base+"/notification-sound", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set(sessionHeader, e.s.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func soundIsCustom(t *testing.T, e *testEnv) bool {
	return e.do(t, http.MethodHead, "/notification-sound", "", nil).Header.Get("X-Custom-Sound") == "1"
}

func TestSoundUploadRejectsUnsupportedWithoutFakeSuccess(t *testing.T) {
	e := newTestEnv(t)
	if r := soundUpload(t, e, bytes.Repeat([]byte("OggS"), 64)); r.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", r.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(e.dir, "notification.wav")); err == nil {
		t.Error("custom WAV must not exist after a rejected upload")
	}
	if soundIsCustom(t, e) {
		t.Error("HEAD reports a custom sound after a rejected upload")
	}
}

func TestSoundUploadMP3BecomesCustomWAV(t *testing.T) {
	e := newTestEnv(t)
	if r := soundUpload(t, e, assets.RingtoneMP3()); r.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", r.StatusCode)
	}
	wav, err := os.ReadFile(filepath.Join(e.dir, "notification.wav"))
	if err != nil || string(wav[0:4]) != "RIFF" {
		t.Fatalf("custom WAV not written: %v", err)
	}
	if !soundIsCustom(t, e) {
		t.Error("HEAD does not report the uploaded custom sound")
	}
	e.do(t, http.MethodDelete, "/notification-sound", "", nil)
	if soundIsCustom(t, e) {
		t.Error("custom sound still reported after reset")
	}
}

func TestSoundUploadAcceptsWAV(t *testing.T) {
	e := newTestEnv(t)
	if r := soundUpload(t, e, assets.NotificationWAV()); r.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", r.StatusCode)
	}
	got, _ := os.ReadFile(filepath.Join(e.dir, "notification.wav"))
	if !bytes.Equal(got, assets.NotificationWAV()) {
		t.Error("uploaded WAV was not stored verbatim")
	}
}
