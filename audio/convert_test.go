package audio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// A non-MP3 input must error, not panic.
func TestConvertMP3ToWAVRejectsGarbage(t *testing.T) {
	if _, err := ConvertMP3ToWAV([]byte("not an mp3 file")); err == nil {
		t.Error("expected error for non-MP3 input")
	}
}

// wrapWAV must produce a canonical, parseable PCM WAV header.
func TestWrapWAVHeader(t *testing.T) {
	pcm := make([]byte, 100)
	w := wrapWAV(pcm, 22050, 1, 16)

	if string(w[0:4]) != "RIFF" || string(w[8:12]) != "WAVE" || string(w[12:16]) != "fmt " {
		t.Fatalf("bad RIFF/WAVE/fmt markers")
	}
	if got := binary.LittleEndian.Uint32(w[4:8]); got != uint32(36+len(pcm)) {
		t.Errorf("RIFF chunk size = %d, want %d", got, 36+len(pcm))
	}
	if got := binary.LittleEndian.Uint16(w[22:24]); got != 1 {
		t.Errorf("channels = %d, want 1", got)
	}
	if got := binary.LittleEndian.Uint32(w[24:28]); got != 22050 {
		t.Errorf("sample rate = %d, want 22050", got)
	}
	di := bytes.Index(w, []byte("data"))
	if di < 0 {
		t.Fatal("no data chunk")
	}
	if got := binary.LittleEndian.Uint32(w[di+4 : di+8]); got != uint32(len(pcm)) {
		t.Errorf("data size = %d, want %d", got, len(pcm))
	}
}
