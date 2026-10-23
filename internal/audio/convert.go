package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/hajimehoshi/go-mp3"
)

const (
	targetRate     = 22050
	maxClipSeconds = 30
)

var (
	ErrNoFrames      = errors.New("no decodable MP3 audio frames found")
	ErrUnsupported   = errors.New("unsupported audio format: upload an MP3 or WAV file")
	ErrTooLong       = errors.New("sound is too long: keep it under 30 seconds")
	bitratesMPEG1    = [16]int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0}
	bitratesMPEG2    = [16]int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0}
	sampleRatesMPEG1 = [3]int{44100, 48000, 32000}
)

func ToWAV(data []byte) ([]byte, error) {
	if IsWAV(data) {
		if WAVDuration(data) > maxClipSeconds {
			return nil, ErrTooLong
		}
		return data, nil
	}
	wav, err := ConvertMP3ToWAV(data)
	if errors.Is(err, ErrNoFrames) {
		return nil, ErrUnsupported
	}
	return wav, err
}

func IsWAV(data []byte) bool {
	if len(data) < 44 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return false
	}
	var haveFmt, haveData bool
	for off := 12; off+8 <= len(data); {
		id := string(data[off : off+4])
		size := int(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		body := off + 8
		if size < 0 || body+size > len(data) {
			if id == "data" && haveFmt {
				return true
			}
			return false
		}
		switch id {
		case "fmt ":
			if size < 16 {
				return false
			}
			format := binary.LittleEndian.Uint16(data[body:])
			channels := binary.LittleEndian.Uint16(data[body+2:])
			rate := binary.LittleEndian.Uint32(data[body+4:])
			bits := binary.LittleEndian.Uint16(data[body+14:])
			if (format != 1 && format != 3 && format != 0xFFFE) || channels < 1 || channels > 2 || rate < 8000 || rate > 192000 || bits == 0 || bits > 32 {
				return false
			}
			haveFmt = true
		case "data":
			haveData = true
		}
		off = body + size + size%2
	}
	return haveFmt && haveData
}

func WAVDuration(data []byte) float64 {
	if !IsWAV(data) {
		return 0
	}
	var byteRate uint32
	for off := 12; off+8 <= len(data); {
		id := string(data[off : off+4])
		size := int(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		body := off + 8
		switch id {
		case "fmt ":
			byteRate = binary.LittleEndian.Uint32(data[body+8:])
		case "data":
			if byteRate == 0 {
				return 0
			}
			return float64(min(size, len(data)-body)) / float64(byteRate)
		}
		off = body + size + size%2
	}
	return 0
}

func ConvertMP3ToWAV(mp3Data []byte) ([]byte, error) {
	frames, err := extractFrames(mp3Data)
	if err != nil {
		return nil, err
	}
	dec, err := mp3.NewDecoder(bytes.NewReader(frames))
	if err != nil {
		return nil, err
	}
	rate := dec.SampleRate()
	raw, err := io.ReadAll(io.LimitReader(dec, int64(rate)*4*maxClipSeconds))
	if err != nil && len(raw) == 0 {
		return nil, err
	}

	step := rate / targetRate
	if step < 1 {
		step = 1
	}
	const frame = 4
	var mono bytes.Buffer
	for i := 0; i+frame <= len(raw); i += frame * step {
		l := int16(binary.LittleEndian.Uint16(raw[i:]))
		r := int16(binary.LittleEndian.Uint16(raw[i+2:]))
		var b [2]byte
		binary.LittleEndian.PutUint16(b[:], uint16(int16((int(l)+int(r))/2)))
		mono.Write(b[:])
	}
	if mono.Len() == 0 {
		return nil, ErrNoFrames
	}

	return wrapWAV(mono.Bytes(), rate/step, 1, 16), nil
}

func extractFrames(b []byte) ([]byte, error) {
	var out bytes.Buffer
	var lockVersion, lockRate byte
	locked := false
	for i := skipID3v2(b); i+4 <= len(b); {
		n := frameLength(b[i:])
		sameStream := !locked || (b[i+1]&0x1E == lockVersion && b[i+2]&0x0C == lockRate)
		if n == 0 || i+n > len(b) || !sameStream || !followedByFrameOrEnd(b, i+n) {
			i++
			continue
		}
		if !locked {
			lockVersion, lockRate, locked = b[i+1]&0x1E, b[i+2]&0x0C, true
		}
		out.Write(b[i : i+n])
		i += n
	}
	if out.Len() == 0 {
		return nil, ErrNoFrames
	}
	return out.Bytes(), nil
}

func followedByFrameOrEnd(b []byte, next int) bool {
	rest := b[next:]
	return len(rest) < 4 || frameLength(rest) > 0 || bytes.HasPrefix(rest, []byte("TAG")) || bytes.HasPrefix(rest, []byte("APETAGEX"))
}

func frameLength(h []byte) int {
	if len(h) < 4 || h[0] != 0xFF || h[1]&0xE0 != 0xE0 {
		return 0
	}
	version := (h[1] >> 3) & 3
	layer := (h[1] >> 1) & 3
	bitrateIdx := h[2] >> 4
	rateIdx := (h[2] >> 2) & 3
	if version == 1 || layer != 1 || bitrateIdx == 0 || bitrateIdx == 15 || rateIdx == 3 {
		return 0
	}
	padding := int(h[2]>>1) & 1
	rate := sampleRatesMPEG1[rateIdx]
	switch version {
	case 3:
		return 144000*bitratesMPEG1[bitrateIdx]/rate + padding
	case 2:
		return 72000*bitratesMPEG2[bitrateIdx]/(rate/2) + padding
	default:
		return 72000*bitratesMPEG2[bitrateIdx]/(rate/4) + padding
	}
}

func skipID3v2(b []byte) int {
	if len(b) < 10 || string(b[:3]) != "ID3" {
		return 0
	}
	size := int(b[6]&0x7F)<<21 | int(b[7]&0x7F)<<14 | int(b[8]&0x7F)<<7 | int(b[9]&0x7F)
	n := 10 + size
	if b[5]&0x10 != 0 {
		n += 10
	}
	if n > len(b) {
		return len(b)
	}
	return n
}

func wrapWAV(pcm []byte, sampleRate, channels, bits int) []byte {
	var buf bytes.Buffer
	w := func(v interface{}) { _ = binary.Write(&buf, binary.LittleEndian, v) }
	byteRate := sampleRate * channels * bits / 8
	buf.WriteString("RIFF")
	w(uint32(36 + len(pcm)))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	w(uint32(16))
	w(uint16(1))
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
