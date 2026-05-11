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

// customPNG holds the raw favicon PNG bytes for blending with status colors.
var customPNG []byte

// Pre-generated blended icons (favicon + colored background), created once in SetCustomIcon.
var (
	iconGreen  []byte
	iconYellow []byte
	iconRed    []byte
	iconGray   []byte
)

// SetCustomIcon blends the favicon with each status color and caches the 4 variants.
func SetCustomIcon(pngData []byte) {
	if len(pngData) == 0 {
		return
	}
	customPNG = pngData
	iconGreen  = blendOnColor(34, 197, 94)
	iconYellow = blendOnColor(234, 179, 8)
	iconRed    = blendOnColor(239, 68, 68)
	iconGray   = blendOnColor(148, 163, 184)
}

func GreenIcon()  []byte { return orFallback(iconGreen,  34, 197, 94) }
func YellowIcon() []byte { return orFallback(iconYellow, 234, 179, 8) }
func RedIcon()    []byte { return orFallback(iconRed,    239, 68, 68) }
func GrayIcon()   []byte { return orFallback(iconGray,   148, 163, 184) }

func orFallback(cached []byte, r, g, b uint8) []byte {
	if len(cached) > 0 {
		return cached
	}
	return generateFallback(r, g, b)
}

// blendOnColor draws a colored circle, then places the favicon (70% size) centered on it.
// The colored ring around the favicon is always visible regardless of favicon transparency.
func blendOnColor(r, g, b uint8) []byte {
	const size = 32
	const iconSize = 22                    // favicon rendered at 22x22 (≈70%)
	const offset = (size - iconSize) / 2  // centered: 5px border

	// Decode favicon PNG
	src, err := png.Decode(bytes.NewReader(customPNG))
	if err != nil {
		return generateFallback(r, g, b)
	}

	dst := image.NewNRGBA(image.Rect(0, 0, size, size))

	// 1. Full colored circle background
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

	// 2. White inner circle (background for the favicon)
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

	// 3. Scale favicon to iconSize × iconSize (nearest-neighbor)
	srcB := src.Bounds()
	scaled := image.NewNRGBA(image.Rect(0, 0, iconSize, iconSize))
	for y := 0; y < iconSize; y++ {
		for x := 0; x < iconSize; x++ {
			sx := srcB.Min.X + x*srcB.Dx()/iconSize
			sy := srcB.Min.Y + y*srcB.Dy()/iconSize
			scaled.Set(x, y, src.At(sx, sy))
		}
	}

	// 4. Draw favicon centered inside the white circle
	iconRect := image.Rect(offset, offset, offset+iconSize, offset+iconSize)
	draw.Draw(dst, iconRect, scaled, image.Point{}, draw.Over)

	var buf bytes.Buffer
	png.Encode(&buf, dst)
	return pngToICO(buf.Bytes())
}

// generateFallback draws a plain 32x32 colored circle (used when no favicon is set).
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

// pngToICO wraps any PNG in a minimal ICO container (PNG-in-ICO, Vista+).
func pngToICO(pngData []byte) []byte {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, uint16(0)) // reserved
	binary.Write(buf, binary.LittleEndian, uint16(1)) // type = ICO
	binary.Write(buf, binary.LittleEndian, uint16(1)) // count = 1
	buf.WriteByte(0)                                                  // width  (0 = from PNG)
	buf.WriteByte(0)                                                  // height (0 = from PNG)
	buf.WriteByte(0)                                                  // color count
	buf.WriteByte(0)                                                  // reserved
	binary.Write(buf, binary.LittleEndian, uint16(1))                 // planes
	binary.Write(buf, binary.LittleEndian, uint16(32))                // bpp
	binary.Write(buf, binary.LittleEndian, uint32(len(pngData)))      // image size
	binary.Write(buf, binary.LittleEndian, uint32(6+16))              // offset
	buf.Write(pngData)
	return buf.Bytes()
}
