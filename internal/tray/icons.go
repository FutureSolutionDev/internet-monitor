package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
)

var customPNG []byte

var (
	iconGreen  []byte
	iconYellow []byte
	iconRed    []byte
	iconGray   []byte
)

func SetCustomIcon(pngData []byte) {
	if len(pngData) == 0 {
		return
	}
	customPNG = pngData
	iconGreen = blendOnColor(34, 197, 94)
	iconYellow = blendOnColor(234, 179, 8)
	iconRed = blendOnColor(239, 68, 68)
	iconGray = blendOnColor(148, 163, 184)
}

func GreenIcon() []byte  { return orFallback(iconGreen, 34, 197, 94) }
func YellowIcon() []byte { return orFallback(iconYellow, 234, 179, 8) }
func RedIcon() []byte    { return orFallback(iconRed, 239, 68, 68) }
func GrayIcon() []byte   { return orFallback(iconGray, 148, 163, 184) }

func orFallback(cached []byte, r, g, b uint8) []byte {
	if len(cached) > 0 {
		return cached
	}
	return generateFallback(r, g, b)
}

func blendOnColor(r, g, b uint8) []byte {
	data, ok := badgePNG(32, 22, r, g, b)
	if !ok {
		return generateFallback(r, g, b)
	}
	return pngToICO(data)
}

func badgePNG(size, iconSize int, r, g, b uint8) ([]byte, bool) {
	offset := (size - iconSize) / 2

	src, err := png.Decode(bytes.NewReader(customPNG))
	if err != nil {
		return nil, false
	}

	dst := image.NewNRGBA(image.Rect(0, 0, size, size))

	cx, cy := float64(size)/2.0, float64(size)/2.0
	outerR := cx - 0.5
	bg := color.NRGBA{R: r, G: g, B: b, A: 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx + 0.5
			dy := float64(y) - cy + 0.5
			if math.Sqrt(dx*dx+dy*dy) <= outerR {
				dst.Set(x, y, bg)
			}
		}
	}

	innerR := float64(iconSize)/2.0 + 1.0
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 230}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx + 0.5
			dy := float64(y) - cy + 0.5
			if math.Sqrt(dx*dx+dy*dy) <= innerR {
				dst.Set(x, y, white)
			}
		}
	}

	srcB := src.Bounds()
	scaled := image.NewNRGBA(image.Rect(0, 0, iconSize, iconSize))
	for y := 0; y < iconSize; y++ {
		for x := 0; x < iconSize; x++ {
			sx := srcB.Min.X + x*srcB.Dx()/iconSize
			sy := srcB.Min.Y + y*srcB.Dy()/iconSize
			scaled.Set(x, y, src.At(sx, sy))
		}
	}

	iconRect := image.Rect(offset, offset, offset+iconSize, offset+iconSize)
	draw.Draw(dst, iconRect, scaled, image.Point{}, draw.Over)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, false
	}
	return buf.Bytes(), true
}

func generateFallback(r, g, b uint8) []byte {
	size := 32
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	cx, cy := float64(size)/2.0, float64(size)/2.0
	radius := cx - 2.5

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx + 0.5
			dy := float64(y) - cy + 0.5
			if dist := math.Sqrt(dx*dx + dy*dy); dist <= radius {
				brightness := 1.0 - (dist/radius)*0.2
				img.Set(x, y, color.NRGBA{
					R: uint8(float64(r) * brightness),
					G: uint8(float64(g) * brightness),
					B: uint8(float64(b) * brightness),
					A: 255,
				})
			}
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return pngToICO(buf.Bytes())
}

func pngToICO(pngData []byte) []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, uint16(0))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint16(1))
	buf.WriteByte(0)
	buf.WriteByte(0)
	buf.WriteByte(0)
	buf.WriteByte(0)
	binary.Write(buf, binary.LittleEndian, uint16(1))
	binary.Write(buf, binary.LittleEndian, uint16(32))
	binary.Write(buf, binary.LittleEndian, uint32(len(pngData)))
	binary.Write(buf, binary.LittleEndian, uint32(6+16))
	buf.Write(pngData)
	return buf.Bytes()
}
