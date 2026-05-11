package render

import "math"

// Canvas is a float intensity grid in [0,1].
//
// PixelAspect = visual-pixel-height / visual-pixel-width.
//   1.0 = square pixels (image output)
//   2.0 = terminal cells (taller than wide)
// Generators use it to keep round shapes round across both targets.
type Canvas struct {
	W, H        int
	PixelAspect float64
	Data        []float64
}

func NewCanvas(w, h int) *Canvas {
	return &Canvas{W: w, H: h, PixelAspect: 1.0, Data: make([]float64, w*h)}
}

func (c *Canvas) Set(x, y int, v float64) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	c.Data[y*c.W+x] = v
}

func (c *Canvas) Add(x, y int, v float64) {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return
	}
	d := c.Data[y*c.W+x] + v
	if d > 1 {
		d = 1
	}
	c.Data[y*c.W+x] = d
}

func (c *Canvas) Get(x, y int) float64 {
	if x < 0 || y < 0 || x >= c.W || y >= c.H {
		return 0
	}
	return c.Data[y*c.W+x]
}

// Normalize stretches values to [0,1].
func (c *Canvas) Normalize() {
	min, max := math.Inf(1), math.Inf(-1)
	for _, v := range c.Data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if max-min < 1e-9 {
		return
	}
	span := max - min
	for i, v := range c.Data {
		c.Data[i] = (v - min) / span
	}
}
