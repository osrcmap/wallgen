package generator

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"golang.org/x/image/draw"

	"wallgen/internal/config"
	"wallgen/internal/render"
)

// Image decodes path, resizes to canvas dims (correcting cell aspect 2:1),
// supersamples by quality, and writes intensity into a Canvas.
func Image(path string, w, h int, pixelAspect float64, q config.Quality, invert bool) (*render.Canvas, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	samples := q.Samples()
	tw, th := w*samples, h*samples
	dst := image.NewGray(image.Rect(0, 0, tw, th))

	var scaler draw.Scaler = draw.CatmullRom
	if q == config.QLow {
		scaler = draw.ApproxBiLinear
	}
	scaler.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	c := render.NewCanvas(w, h)
	if pixelAspect > 0 {
		c.PixelAspect = pixelAspect
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sum float64
			var n int
			for sy := 0; sy < samples; sy++ {
				for sx := 0; sx < samples; sx++ {
					px := x*samples + sx
					py := y*samples + sy
					g := dst.GrayAt(px, py).Y
					sum += float64(g) / 255.0
					n++
				}
			}
			v := sum / float64(n)
			if invert {
				v = 1 - v
			}
			c.Set(x, y, v)
		}
	}
	return c, nil
}
