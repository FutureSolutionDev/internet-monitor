package sound

import (
	"encoding/binary"
	"math"
	"testing"
)

func wav(code, bits uint16, pcm []byte, extra ...byte) []byte {
	fmtBody := make([]byte, 16)
	binary.LittleEndian.PutUint16(fmtBody[0:], code)
	binary.LittleEndian.PutUint16(fmtBody[2:], 1)
	binary.LittleEndian.PutUint32(fmtBody[4:], 8000)
	binary.LittleEndian.PutUint16(fmtBody[14:], bits)
	var b []byte
	b = append(b, "RIFF\x00\x00\x00\x00WAVE"...)
	b = append(b, "fmt "...)
	b = binary.LittleEndian.AppendUint32(b, uint32(len(fmtBody)))
	b = append(b, fmtBody...)
	if len(extra) > 0 {
		b = append(b, "LIST"...)
		b = binary.LittleEndian.AppendUint32(b, uint32(len(extra)))
		b = append(b, extra...)
		if len(extra)%2 == 1 {
			b = append(b, 0)
		}
	}
	b = append(b, "data"...)
	b = binary.LittleEndian.AppendUint32(b, uint32(len(pcm)))
	b = append(b, pcm...)
	binary.LittleEndian.PutUint32(b[4:], uint32(len(b)-8))
	return b
}

func samples16(vals ...int16) []byte {
	var b []byte
	for _, v := range vals {
		b = binary.LittleEndian.AppendUint16(b, uint16(v))
	}
	return b
}

func read16(t *testing.T, data []byte) []int16 {
	t.Helper()
	f, ok := parseWAV(data)
	if !ok {
		t.Fatal("scaled output is not a WAV")
	}
	var out []int16
	for i := f.dataStart; i+2 <= f.dataEnd; i += 2 {
		out = append(out, int16(binary.LittleEndian.Uint16(data[i:])))
	}
	return out
}

func TestGainCurve(t *testing.T) {
	cases := map[int]float64{-5: 0, 0: 0, 50: 0.25, 100: 1, 150: 1}
	for in, want := range cases {
		if got := Gain(in); math.Abs(got-want) > 1e-9 {
			t.Errorf("Gain(%d) = %v, want %v", in, got, want)
		}
	}
}

func TestScale16BitWithoutOverflow(t *testing.T) {
	src := wav(wavPCM, 16, samples16(math.MaxInt16, math.MinInt16, 1000, -1000, 0), 'a', 'b', 'c')
	out, ok := ScaleWAV(src, 50)
	if !ok {
		t.Fatal("16-bit PCM was not scaled")
	}
	got := read16(t, out)
	want := []int16{8192, -8192, 250, -250, 0}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("samples = %v, want %v", got, want)
		}
	}
	if read16(t, src)[0] != math.MaxInt16 {
		t.Fatal("source buffer was modified")
	}
	assertConsistentHeader(t, out)
}

func assertConsistentHeader(t *testing.T, out []byte) {
	t.Helper()
	if riff := int(binary.LittleEndian.Uint32(out[4:])); riff != len(out)-8 {
		t.Fatalf("RIFF size %d, file %d", riff, len(out))
	}
	f, ok := parseWAV(out)
	if !ok {
		t.Fatal("not a WAV")
	}
	if declared := int(binary.LittleEndian.Uint32(out[f.dataStart-4:])); declared != f.dataEnd-f.dataStart || f.dataEnd > len(out) {
		t.Fatalf("data size %d, actual %d", declared, f.dataEnd-f.dataStart)
	}
}

func TestScaleRewritesLyingSizes(t *testing.T) {
	src := wav(wavPCM, 16, samples16(1000, 2000, 3000))
	binary.LittleEndian.PutUint32(src[4:], 0xFFFFFFF0)
	binary.LittleEndian.PutUint32(src[len(src)-6-4:], 1<<30)
	out, ok := ScaleWAV(src, 50)
	if !ok {
		t.Fatal("not scaled")
	}
	assertConsistentHeader(t, out)
	if got := read16(t, out); len(got) != 3 || got[2] != 750 {
		t.Fatalf("samples = %v", got)
	}
}

func TestScaleStreamingDataSizeZero(t *testing.T) {
	src := wav(wavPCM, 16, samples16(400, -400))
	binary.LittleEndian.PutUint32(src[len(src)-4-4:], 0)
	out, ok := ScaleWAV(src, 50)
	if !ok {
		t.Fatal("not scaled")
	}
	assertConsistentHeader(t, out)
	if got := read16(t, out); len(got) != 2 || got[0] != 100 {
		t.Fatalf("samples = %v", got)
	}
}

func TestScaleToSilence(t *testing.T) {
	out, ok := ScaleWAV(wav(wavPCM, 16, samples16(30000, -30000)), 0)
	if !ok {
		t.Fatal("not scaled")
	}
	for _, s := range read16(t, out) {
		if s != 0 {
			t.Fatalf("volume 0 left %d", s)
		}
	}
}

func TestScale8BitIsCentredOn128(t *testing.T) {
	out, ok := ScaleWAV(wav(wavPCM, 8, []byte{255, 0, 128}), 50)
	if !ok {
		t.Fatal("not scaled")
	}
	f, _ := parseWAV(out)
	got := out[f.dataStart:f.dataEnd]
	if got[0] != 160 || got[1] != 96 || got[2] != 128 {
		t.Fatalf("8-bit = %v", got)
	}
}

func TestScaleOtherDepths(t *testing.T) {
	pcm24 := []byte{0xff, 0xff, 0x7f, 0x00, 0x00, 0x80}
	out, ok := ScaleWAV(wav(wavPCM, 24, pcm24), 50)
	if !ok {
		t.Fatal("24-bit not scaled")
	}
	f, _ := parseWAV(out)
	p := out[f.dataStart:]
	if v := int32(uint32(p[0])|uint32(p[1])<<8|uint32(p[2])<<16) << 8 >> 8; v != 2097152 {
		t.Errorf("24-bit max = %d", v)
	}
	if v := int32(uint32(p[3])|uint32(p[4])<<8|uint32(p[5])<<16) << 8 >> 8; v != -2097152 {
		t.Errorf("24-bit min = %d", v)
	}

	fl := binary.LittleEndian.AppendUint32(nil, math.Float32bits(0.8))
	out, ok = ScaleWAV(wav(wavFloat, 32, fl), 50)
	if !ok {
		t.Fatal("float not scaled")
	}
	f, _ = parseWAV(out)
	if v := math.Float32frombits(binary.LittleEndian.Uint32(out[f.dataStart:])); math.Abs(float64(v)-0.2) > 1e-6 {
		t.Errorf("float = %v", v)
	}
}

func TestScaleRejectsUnknownData(t *testing.T) {
	for name, data := range map[string][]byte{
		"empty":     nil,
		"not riff":  []byte("hello world, not a wav"),
		"adpcm":     wav(2, 4, []byte{1, 2, 3}),
		"no fmt":    append([]byte("RIFF\x10\x00\x00\x00WAVE"), "data\x02\x00\x00\x00ab"...),
		"truncated": wav(wavPCM, 16, samples16(1, 2, 3))[:30],
		"empty pcm": wav(wavPCM, 16, nil),
	} {
		if _, ok := ScaleWAV(data, 50); ok {
			t.Errorf("%s: scaled", name)
		}
	}
}
