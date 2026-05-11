package tray

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
)

// pngToICO wraps a PNG byte slice in a valid .ico container.
// getlantern/systray on Windows requires ICO format, not raw PNG.
func pngToICO(pngData []byte) []byte {
	buf := &bytes.Buffer{}

	// ICONDIR header (6 bytes)
	binary.Write(buf, binary.LittleEndian, uint16(0)) // reserved
	binary.Write(buf, binary.LittleEndian, uint16(1)) // type = 1 (icon)
	binary.Write(buf, binary.LittleEndian, uint16(1)) // image count = 1

	// ICONDIRENTRY (16 bytes)
	buf.WriteByte(32)                                                       // width  = 32px
	buf.WriteByte(32)                                                       // height = 32px
	buf.WriteByte(0)                                                        // color count (0 = >256 colors)
	buf.WriteByte(0)                                                        // reserved
	binary.Write(buf, binary.LittleEndian, uint16(1))                      // planes
	binary.Write(buf, binary.LittleEndian, uint16(32))                     // bits per pixel (RGBA)
	binary.Write(buf, binary.LittleEndian, uint32(len(pngData)))           // size of image data
	binary.Write(buf, binary.LittleEndian, uint32(6+16))                   // offset: ICONDIR(6) + ICONDIRENTRY(16)

	buf.Write(pngData)
	return buf.Bytes()
}

func generateIcon(r, g, b uint8) []byte {
	size := 32
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	cx := float64(size) / 2.0
	cy := float64(size) / 2.0
	radius := cx - 2.5

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx + 0.5
			dy := float64(y) - cy + 0.5
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= radius {
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

func GreenIcon()  []byte { return generateIcon(34, 197, 94) }
func YellowIcon() []byte { return generateIcon(234, 179, 8) }
func RedIcon()    []byte { return generateIcon(239, 68, 68) }
func GrayIcon()   []byte { return generateIcon(148, 163, 184) }
