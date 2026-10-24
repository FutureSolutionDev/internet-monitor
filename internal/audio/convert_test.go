package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"testing"
)

func loadDefaultMP3(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../assets/Ringtone.mp3")
	if err != nil {
		t.Fatalf("read default ringtone: %v", err)
	}
	return data
}

func wavSampleRate(w []byte) uint32 { return binary.LittleEndian.Uint32(w[24:28]) }

func wavDataLen(t *testing.T, w []byte) int {
	t.Helper()
	di := bytes.Index(w, []byte("data"))
	if di < 0 {
		t.Fatal("no data chunk")
	}
	return int(binary.LittleEndian.Uint32(w[di+4 : di+8]))
}

func TestConvertMP3ToWAVRejectsGarbage(t *testing.T) {
	if _, err := ConvertMP3ToWAV([]byte("not an mp3 file")); err == nil {
		t.Error("expected error for non-MP3 input")
	}
}

func TestConvertDefaultRingtone(t *testing.T) {
	w, err := ConvertMP3ToWAV(loadDefaultMP3(t))
	if err != nil {
		t.Fatalf("convert default ringtone: %v", err)
	}
	if !IsWAV(w) {
		t.Fatal("output is not a WAV")
	}
	if got := wavSampleRate(w); got != 22050 {
		t.Errorf("sample rate = %d, want 22050", got)
	}
	if n := wavDataLen(t, w); n < 22050 {
		t.Errorf("data too short: %d bytes", n)
	}
}

func TestConvertSurvivesJunkBetweenFrames(t *testing.T) {
	src := loadDefaultMP3(t)
	clean, err := ConvertMP3ToWAV(src)
	if err != nil {
		t.Fatalf("baseline convert: %v", err)
	}

	start := skipID3v2(src)
	junk := []byte{0xFF, 0xFF, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, 0x12, 0x34}
	first := frameLength(src[start:])
	if first == 0 {
		t.Fatal("default ringtone has no frame right after ID3")
	}
	var dirty []byte
	dirty = append(dirty, src[:start]...)
	dirty = append(dirty, junk...)
	dirty = append(dirty, src[start:start+first]...)
	dirty = append(dirty, junk...)
	dirty = append(dirty, src[start+first:]...)

	w, err := ConvertMP3ToWAV(dirty)
	if err != nil {
		t.Fatalf("convert with junk: %v", err)
	}
	const oneFrame = 1152
	if got, want := wavDataLen(t, w), wavDataLen(t, clean); got > want || want-got > oneFrame {
		t.Errorf("data length with junk = %d, want within one frame of %d", got, want)
	}
}

func TestToWAVPassesThroughWAV(t *testing.T) {
	in := wrapWAV(make([]byte, 200), 44100, 1, 16)
	out, err := ToWAV(in)
	if err != nil {
		t.Fatalf("ToWAV(wav): %v", err)
	}
	if !bytes.Equal(in, out) {
		t.Error("WAV input was modified")
	}
}

func TestToWAVRejectsUnknownFormat(t *testing.T) {
	_, err := ToWAV(bytes.Repeat([]byte("OggS"), 64))
	if !errors.Is(err, ErrUnsupported) {
		t.Errorf("err = %v, want ErrUnsupported", err)
	}
}

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
	if got := wavSampleRate(w); got != 22050 {
		t.Errorf("sample rate = %d, want 22050", got)
	}
	if got := wavDataLen(t, w); got != len(pcm) {
		t.Errorf("data size = %d, want %d", got, len(pcm))
	}
}
