// Package audio converts MP3 -> WAV for user-uploaded custom sounds.
//
// The native player (winmm PlaySound) is WAV-only, so an uploaded MP3 is
// transcoded to a small mono 22 kHz 16-bit PCM WAV. go-mp3 is a pure-Go
// decoder, so this adds no CGO and doesn't break cross-compilation. This
// package has no internal deps, so dashboard can import it without a cycle.
package audio

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/hajimehoshi/go-mp3"
)

// ConvertMP3ToWAV decodes MP3 bytes and returns a mono 22050 Hz 16-bit PCM WAV.
func ConvertMP3ToWAV(mp3Data []byte) ([]byte, error) {
	dec, err := mp3.NewDecoder(bytes.NewReader(mp3Data))
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(dec) // 16-bit LE stereo @ dec.SampleRate()
	if err != nil {
		return nil, err
	}

	const dstRate = 22050
	step := dec.SampleRate() / dstRate
	if step < 1 {
		step = 1
	}
	const frame = 4 // stereo 16-bit frame
	var mono bytes.Buffer
	for i := 0; i+frame <= len(raw); i += frame * step {
		l := int16(binary.LittleEndian.Uint16(raw[i:]))
		r := int16(binary.LittleEndian.Uint16(raw[i+2:]))
		var b [2]byte
		binary.LittleEndian.PutUint16(b[:], uint16(int16((int(l)+int(r))/2)))
		mono.Write(b[:])
	}

	return wrapWAV(mono.Bytes(), dstRate, 1, 16), nil
}

// wrapWAV prepends a canonical PCM WAV header to raw little-endian samples.
func wrapWAV(pcm []byte, sampleRate, channels, bits int) []byte {
	var buf bytes.Buffer
	w := func(v interface{}) { binary.Write(&buf, binary.LittleEndian, v) }
	byteRate := sampleRate * channels * bits / 8
	buf.WriteString("RIFF")
	w(uint32(36 + len(pcm)))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	w(uint32(16))
	w(uint16(1)) // PCM
	w(uint16(channels))
	w(uint32(sampleRate))
	w(uint32(byteRate))
	w(uint16(channels * bits / 8))
	w(uint16(bits))
	buf.WriteString("data")
	w(uint32(len(pcm)))
	buf.Write(pcm)
	return buf.Bytes()
}
