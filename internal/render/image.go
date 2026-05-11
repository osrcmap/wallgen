package render

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"strings"
)

type Palette string

const (
	PalMono    Palette = "mono"
	PalInverse Palette = "inverse"
	PalFire    Palette = "fire"
	PalOcean   Palette = "ocean"
	PalViridis Palette = "viridis"
	PalRainbow Palette = "rainbow"
	PalSunset  Palette = "sunset"
)

var Palettes = []Palette{PalMono, PalInverse, PalFire, PalOcean, PalViridis, PalRainbow, PalSunset}

// Sample returns the RGBA color for a normalized intensity v in [0,1].
func (p Palette) Sample(v float64) color.RGBA {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	switch p {
	case PalInverse:
		g := uint8(255 - v*255)
		return color.RGBA{g, g, g, 255}
	case PalFire:
		return gradient(v, color.RGBA{0, 0, 0, 255}, color.RGBA{120, 0, 0, 255}, color.RGBA{255, 80, 0, 255}, color.RGBA{255, 220, 60, 255}, color.RGBA{255, 255, 220, 255})
	case PalOcean:
		return gradient(v, color.RGBA{2, 4, 30, 255}, color.RGBA{8, 30, 90, 255}, color.RGBA{20, 90, 160, 255}, color.RGBA{80, 200, 220, 255}, color.RGBA{220, 250, 255, 255})
	case PalViridis:
		return gradient(v, color.RGBA{68, 1, 84, 255}, color.RGBA{59, 82, 139, 255}, color.RGBA{33, 144, 140, 255}, color.RGBA{93, 201, 99, 255}, color.RGBA{253, 231, 37, 255})
	case PalRainbow:
		h := v * 300
		return hsv(h, 0.85, 0.95)
	case PalSunset:
		return gradient(v, color.RGBA{20, 10, 40, 255}, color.RGBA{120, 30, 100, 255}, color.RGBA{220, 90, 90, 255}, color.RGBA{255, 180, 80, 255}, color.RGBA{255, 240, 200, 255})
	default:
		g := uint8(v * 255)
		return color.RGBA{g, g, g, 255}
	}
}

func gradient(v float64, stops ...color.RGBA) color.RGBA {
	n := len(stops)
	if n < 2 {
		return stops[0]
	}
	pos := v * float64(n-1)
	i := int(pos)
	if i >= n-1 {
		return stops[n-1]
	}
	t := pos - float64(i)
	a, b := stops[i], stops[i+1]
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: 255,
	}
}

func hsv(h, s, v float64) color.RGBA {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return color.RGBA{uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255), 255}
}

// EncodeImage writes the canvas as PNG or JPG to w. Each canvas cell becomes
// `scale` x `scale` pixels.
func EncodeImage(c *Canvas, p Palette, scale int, format string, w io.Writer) error {
	if scale < 1 {
		scale = 1
	}
	pw, ph := c.W*scale, c.H*scale
	img := image.NewRGBA(image.Rect(0, 0, pw, ph))
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			col := p.Sample(c.Get(x, y))
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					img.SetRGBA(x*scale+dx, y*scale+dy, col)
				}
			}
		}
	}
	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 92})
	case "png":
		return png.Encode(w, img)
	default:
		return fmt.Errorf("unsupported image format %q", format)
	}
}

func ParsePalette(s string) Palette {
	for _, p := range Palettes {
		if string(p) == strings.ToLower(s) {
			return p
		}
	}
	return PalMono
}
