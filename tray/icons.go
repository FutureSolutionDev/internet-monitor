package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
)

// customICO holds the user-supplied favicon wrapped in ICO format.
// Set via SetCustomIcon before systray.Run() is called.
var customICO []byte

// SetCustomIcon converts a PNG byte slice into an ICO and uses it for all tray states.
func SetCustomIcon(pngData []byte) {
	if len(pngData) > 0 {
		customICO = pngToICO(pngData)
	}
}

// pngToICO wraps any PNG in a minimal valid ICO container.
// Width/Height = 0 in the directory entry tells Windows to read size from the PNG header.
// This works for all sizes on Windows Vista+ (the PNG-in-ICO format).
func pngToICO(pngData []byte) []byte {
	buf := &bytes.Buffer{}
	// ICONDIR (6 bytes)
	binary.Write(buf, binary.LittleEndian, uint16(0)) // reserved
	binary.Write(buf, binary.LittleEndian, uint16(1)) // type = ICO
	binary.Write(buf, binary.LittleEndian, uint16(1)) // image count
	// ICONDIRENTRY (16 bytes)
	buf.WriteByte(0)                                                       // width  — 0 = read from PNG
	buf.WriteByte(0)                                                       // height — 0 = read from PNG
	buf.WriteByte(0)                                                       // color count
	buf.WriteByte(0)                                                       // reserved
	binary.Write(buf, binary.LittleEndian, uint16(1))                     // planes
	binary.Write(buf, binary.LittleEndian, uint16(32))                    // bits per pixel
	binary.Write(buf, binary.LittleEndian, uint32(len(pngData)))          // size of image
	binary.Write(buf, binary.LittleEndian, uint32(6+16))                  // offset (ICONDIR + ICONDIRENTRY)
	buf.Write(pngData)
	return buf.Bytes()
}

func icon(r, g, b uint8) []byte {
	if customICO != nil {
		return customICO
	}
	return generateFallback(r, g, b)
}

func GreenIcon()  []byte { return icon(34, 197, 94) }
func YellowIcon() []byte { return icon(234, 179, 8) }
func RedIcon()    []byte { return icon(239, 68, 68) }
func GrayIcon()   []byte { return icon(148, 163, 184) }

// generateFallback draws a 32x32 colored circle when no custom icon is set.
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
