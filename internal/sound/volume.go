package sound

import (
	"encoding/binary"
	"math"
)

const (
	wavPCM        = 1
	wavFloat      = 3
	wavExtensible = 0xFFFE
)

func Gain(volume int) float64 {
	if volume <= 0 {
		return 0
	}
	if volume >= 100 {
		return 1
	}
	v := float64(volume) / 100
	return v * v
}

const MaxScaleBytes = 16 << 20

type wavFormat struct {
	code      uint16
	bits      uint16
	fmtStart  int
	fmtEnd    int
	dataStart int
	dataEnd   int
}

func parseWAV(data []byte) (wavFormat, bool) {
	var f wavFormat
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return f, false
	}
	haveFmt := false
	for off := 12; off+8 <= len(data); {
		id := string(data[off : off+4])
		size := int(binary.LittleEndian.Uint32(data[off+4 : off+8]))
		body := off + 8
		end := body + size
		if size < 0 || end > len(data) || end < body || (id == "data" && size == 0) {
			end = len(data)
		}
		switch id {
		case "fmt ":
			if end-body < 16 {
				return f, false
			}
			f.fmtStart, f.fmtEnd = body, end
			f.code = binary.LittleEndian.Uint16(data[body : body+2])
			f.bits = binary.LittleEndian.Uint16(data[body+14 : body+16])
			if f.code == wavExtensible && end-body >= 26 {
				f.code = binary.LittleEndian.Uint16(data[body+24 : body+26])
			}
			haveFmt = true
		case "data":
			if !haveFmt {
				return f, false
			}
			f.dataStart, f.dataEnd = body, end
			return f, true
		}
		off = end + (size & 1)
	}
	return f, false
}

func ScaleWAV(data []byte, volume int) ([]byte, bool) {
	if len(data) > MaxScaleBytes {
		return nil, false
	}
	f, ok := parseWAV(data)
	if !ok || f.dataEnd <= f.dataStart {
		return nil, false
	}
	gain := Gain(volume)
	out, pcm := normalizedWAV(data, f)
	switch {
	case f.code == wavPCM && f.bits == 8:
		for i, b := range pcm {
			pcm[i] = byte(clampInt(int(math.Round(float64(int(b)-128)*gain))+128, 0, 255))
		}
	case f.code == wavPCM && f.bits == 16:
		for i := 0; i+2 <= len(pcm); i += 2 {
			s := float64(int16(binary.LittleEndian.Uint16(pcm[i:])))
			binary.LittleEndian.PutUint16(pcm[i:], uint16(int16(clampInt(int(math.Round(s*gain)), math.MinInt16, math.MaxInt16))))
		}
	case f.code == wavPCM && f.bits == 24:
		for i := 0; i+3 <= len(pcm); i += 3 {
			s := int32(uint32(pcm[i])|uint32(pcm[i+1])<<8|uint32(pcm[i+2])<<16) << 8 >> 8
			v := clampInt(int(math.Round(float64(s)*gain)), -1<<23, 1<<23-1)
			pcm[i], pcm[i+1], pcm[i+2] = byte(v), byte(v>>8), byte(v>>16)
		}
	case f.code == wavPCM && f.bits == 32:
		for i := 0; i+4 <= len(pcm); i += 4 {
			s := float64(int32(binary.LittleEndian.Uint32(pcm[i:])))
			binary.LittleEndian.PutUint32(pcm[i:], uint32(int32(clampInt(int(math.Round(s*gain)), math.MinInt32, math.MaxInt32))))
		}
	case f.code == wavFloat && f.bits == 32:
		for i := 0; i+4 <= len(pcm); i += 4 {
			s := math.Float32frombits(binary.LittleEndian.Uint32(pcm[i:]))
			binary.LittleEndian.PutUint32(pcm[i:], math.Float32bits(s*float32(gain)))
		}
	default:
		return nil, false
	}
	return out, true
}

func normalizedWAV(data []byte, f wavFormat) (out, pcm []byte) {
	fmtBody := data[f.fmtStart:f.fmtEnd]
	pcmLen := f.dataEnd - f.dataStart
	size := 12 + 8 + len(fmtBody) + len(fmtBody)&1 + 8 + pcmLen + pcmLen&1
	out = make([]byte, 0, size)
	out = append(out, "RIFF"...)
	out = binary.LittleEndian.AppendUint32(out, uint32(size-8))
	out = append(out, "WAVEfmt "...)
	out = binary.LittleEndian.AppendUint32(out, uint32(len(fmtBody)))
	out = append(out, fmtBody...)
	if len(fmtBody)&1 == 1 {
		out = append(out, 0)
	}
	out = append(out, "data"...)
	out = binary.LittleEndian.AppendUint32(out, uint32(pcmLen))
	start := len(out)
	out = append(out, data[f.dataStart:f.dataEnd]...)
	if pcmLen&1 == 1 {
		out = append(out, 0)
	}
	return out, out[start : start+pcmLen]
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
