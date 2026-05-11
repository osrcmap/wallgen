package generator

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"

	"wallgen/internal/config"
	"wallgen/internal/render"
)

type MathKind string

const (
	Mandelbrot   MathKind = "mandelbrot"
	Julia        MathKind = "julia"
	BurningShip  MathKind = "burning_ship"
	Newton       MathKind = "newton"
	Sierpinski   MathKind = "sierpinski"
	Barnsley     MathKind = "barnsley"
	Lissajous    MathKind = "lissajous"
	Spiral       MathKind = "spiral"
	Rose         MathKind = "rose"
	Phyllotaxis  MathKind = "phyllotaxis"
	Interference MathKind = "interference"
	Plasma       MathKind = "plasma"
	Voronoi      MathKind = "voronoi"
	Truchet      MathKind = "truchet"
)

var MathKinds = []MathKind{
	Mandelbrot, Julia, BurningShip, Newton,
	Sierpinski, Barnsley,
	Lissajous, Spiral, Rose, Phyllotaxis,
	Interference, Plasma, Voronoi, Truchet,
}

func Math(kind MathKind, w, h int, pixelAspect float64, q config.Quality) (*render.Canvas, error) {
	c := render.NewCanvas(w, h)
	if pixelAspect > 0 {
		c.PixelAspect = pixelAspect
	}
	switch kind {
	case Mandelbrot:
		mandelbrot(c, q.Iterations())
	case Julia:
		julia(c, q.Iterations())
	case BurningShip:
		burningShip(c, q.Iterations())
	case Newton:
		newton(c, q)
	case Sierpinski:
		sierpinski(c, q)
	case Barnsley:
		barnsley(c, q)
	case Lissajous:
		lissajous(c, q)
	case Spiral:
		spiral(c, q)
	case Rose:
		rose(c, q)
	case Phyllotaxis:
		phyllotaxis(c, q)
	case Interference:
		interference(c, q)
	case Plasma:
		plasma(c, q)
	case Voronoi:
		voronoi(c, q)
	case Truchet:
		truchet(c, q)
	default:
		return nil, fmt.Errorf("unknown math kind %q", kind)
	}
	return c, nil
}

func mandelbrot(c *render.Canvas, maxIter int) {
	cx0, cy0, scale := -0.5, 0.0, 3.2
	yScale := scale * float64(c.H) * c.PixelAspect / float64(c.W)
	for py := 0; py < c.H; py++ {
		for px := 0; px < c.W; px++ {
			x0 := cx0 + (float64(px)/float64(c.W)-0.5)*scale
			y0 := cy0 + (float64(py)/float64(c.H)-0.5)*yScale
			z := complex(0, 0)
			cc := complex(x0, y0)
			var i int
			for i = 0; i < maxIter; i++ {
				if cmplx.Abs(z) > 2 {
					break
				}
				z = z*z + cc
			}
			if i == maxIter {
				c.Set(px, py, 0)
				continue
			}
			// smooth coloring
			nu := math.Log(math.Log(cmplx.Abs(z))/math.Log(2)) / math.Log(2)
			v := (float64(i) + 1 - nu) / float64(maxIter)
			c.Set(px, py, math.Sqrt(v))
		}
	}
}

func julia(c *render.Canvas, maxIter int) {
	cc := complex(-0.7269, 0.1889)
	scale := 3.0
	yScale := scale * float64(c.H) * c.PixelAspect / float64(c.W)
	for py := 0; py < c.H; py++ {
		for px := 0; px < c.W; px++ {
			x := (float64(px)/float64(c.W) - 0.5) * scale
			y := (float64(py)/float64(c.H) - 0.5) * yScale
			z := complex(x, y)
			var i int
			for i = 0; i < maxIter; i++ {
				if cmplx.Abs(z) > 2 {
					break
				}
				z = z*z + cc
			}
			if i == maxIter {
				c.Set(px, py, 1)
				continue
			}
			nu := math.Log(math.Log(cmplx.Abs(z))/math.Log(2)) / math.Log(2)
			v := (float64(i) + 1 - nu) / float64(maxIter)
			c.Set(px, py, math.Pow(v, 0.6))
		}
	}
}

func sierpinski(c *render.Canvas, q config.Quality) {
	// equilateral triangle inscribed in the canvas, accounting for pixel aspect.
	pa := c.PixelAspect
	visualW := float64(c.W)
	visualH := float64(c.H) * pa
	base := visualW
	height := base * math.Sqrt(3) / 2
	if height > visualH {
		height = visualH
		base = height * 2 / math.Sqrt(3)
	}
	cx := visualW / 2
	yTopPix := (visualH - height) / 2 / pa
	yBotPix := yTopPix + height/pa
	verts := [3][2]float64{
		{cx, yTopPix},
		{cx - base/2, yBotPix},
		{cx + base/2, yBotPix},
	}
	r := rand.New(rand.NewSource(42))
	x, y := cx, (yTopPix+yBotPix)/2
	n := c.W * c.H * (q.Samples() + 2)
	for i := 0; i < n; i++ {
		v := verts[r.Intn(3)]
		x = (x + v[0]) / 2
		y = (y + v[1]) / 2
		if i > 20 {
			c.Add(int(x), int(y), 0.6)
		}
	}
	c.Normalize()
}

func lissajous(c *render.Canvas, q config.Quality) {
	a, b, delta := 5.0, 4.0, math.Pi/2
	steps := 20000 * (q.Samples() + 1)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps) * 2 * math.Pi * 3
		x := math.Sin(a*t+delta)
		y := math.Sin(b * t)
		px := int((x + 1) / 2 * float64(c.W-1))
		py := int((y + 1) / 2 * float64(c.H-1))
		c.Add(px, py, 0.15)
	}
}

func spiral(c *render.Canvas, q config.Quality) {
	cx, cy := float64(c.W)/2, float64(c.H)/2
	pa := c.PixelAspect
	maxR := math.Min(cx, cy*pa) * 0.95
	turns := 6.0 + 2*float64(q)
	steps := 30000 * (q.Samples() + 1)
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps)
		r := t * maxR
		theta := t * turns * 2 * math.Pi
		x := cx + r*math.Cos(theta)
		y := cy + r*math.Sin(theta)/pa
		c.Add(int(x), int(y), 0.2)
	}
}

func interference(c *render.Canvas, q config.Quality) {
	sources := [][2]float64{
		{float64(c.W) * 0.3, float64(c.H) * 0.5},
		{float64(c.W) * 0.7, float64(c.H) * 0.5},
		{float64(c.W) * 0.5, float64(c.H) * 0.2},
	}
	pa := c.PixelAspect
	k := 0.5 + 0.2*float64(q)
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			var sum float64
			for _, s := range sources {
				dx := float64(x) - s[0]
				dy := (float64(y) - s[1]) * pa
				d := math.Sqrt(dx*dx + dy*dy)
				sum += math.Sin(d * k)
			}
			v := (sum/float64(len(sources)) + 1) / 2
			c.Set(x, y, v)
		}
	}
}

func rose(c *render.Canvas, q config.Quality) {
	cx, cy := float64(c.W)/2, float64(c.H)/2
	pa := c.PixelAspect
	maxR := math.Min(cx, cy*pa) * 0.9
	n, d := 5.0, 4.0
	k := n / d
	steps := 40000 * (q.Samples() + 1)
	for i := 0; i < steps; i++ {
		theta := float64(i) / float64(steps) * 2 * math.Pi * d
		r := math.Cos(k*theta) * maxR
		x := cx + r*math.Cos(theta)
		y := cy + r*math.Sin(theta)/pa
		c.Add(int(x), int(y), 0.25)
	}
}

// burningShip is mandelbrot variant: z = (|Re|+i|Im|)² + c.
func burningShip(c *render.Canvas, maxIter int) {
	cx0, cy0, scale := -0.5, -0.5, 3.5
	yScale := scale * float64(c.H) * c.PixelAspect / float64(c.W)
	for py := 0; py < c.H; py++ {
		for px := 0; px < c.W; px++ {
			x0 := cx0 + (float64(px)/float64(c.W)-0.5)*scale
			y0 := cy0 + (float64(py)/float64(c.H)-0.5)*yScale
			zx, zy := 0.0, 0.0
			i := 0
			for ; i < maxIter; i++ {
				zx2, zy2 := zx*zx, zy*zy
				if zx2+zy2 > 4 {
					break
				}
				zx, zy = zx2-zy2+x0, 2*math.Abs(zx*zy)+y0
			}
			if i == maxIter {
				c.Set(px, py, 0)
				continue
			}
			v := float64(i) / float64(maxIter)
			c.Set(px, py, math.Sqrt(v))
		}
	}
}

// newton solves z³ - 1 = 0 by Newton's method, intensity encodes which root +
// convergence speed. Three basins meet in fractal boundaries.
func newton(c *render.Canvas, q config.Quality) {
	maxIter := q.Iterations() / 8
	if maxIter < 20 {
		maxIter = 20
	}
	if maxIter > 80 {
		maxIter = 80
	}
	roots := [3]complex128{
		complex(1, 0),
		complex(-0.5, math.Sqrt(3)/2),
		complex(-0.5, -math.Sqrt(3)/2),
	}
	scale := 3.0
	yScale := scale * float64(c.H) * c.PixelAspect / float64(c.W)
	const eps = 1e-6
	for py := 0; py < c.H; py++ {
		for px := 0; px < c.W; px++ {
			z := complex(
				(float64(px)/float64(c.W)-0.5)*scale,
				(float64(py)/float64(c.H)-0.5)*yScale,
			)
			i := 0
			for ; i < maxIter; i++ {
				z2 := z * z
				if cmplx.Abs(z2) < 1e-12 {
					break
				}
				z -= (z*z2 - 1) / (3 * z2)
				done := false
				for _, r := range roots {
					if cmplx.Abs(z-r) < eps {
						done = true
						break
					}
				}
				if done {
					break
				}
			}
			rootIdx := 0
			best := math.Inf(1)
			for k, r := range roots {
				d := cmplx.Abs(z - r)
				if d < best {
					best = d
					rootIdx = k
				}
			}
			shade := 1 - float64(i)/float64(maxIter)
			v := (float64(rootIdx) + shade) / 3
			c.Set(px, py, v)
		}
	}
}

// barnsley fern via 4-map IFS. Best on portrait aspect (e.g. 9:16, 1:1).
func barnsley(c *render.Canvas, q config.Quality) {
	const minX, maxX = -2.2, 2.7
	const minY, maxY = 0.0, 10.0
	r := rand.New(rand.NewSource(7))
	x, y := 0.0, 0.0
	n := c.W * c.H * (q.Samples() + 4)
	for i := 0; i < n; i++ {
		rr := r.Float64()
		var nx, ny float64
		switch {
		case rr < 0.01:
			ny = 0.16 * y
		case rr < 0.86:
			nx = 0.85*x + 0.04*y
			ny = -0.04*x + 0.85*y + 1.6
		case rr < 0.93:
			nx = 0.20*x - 0.26*y
			ny = 0.23*x + 0.22*y + 1.6
		default:
			nx = -0.15*x + 0.28*y
			ny = 0.26*x + 0.24*y + 0.44
		}
		x, y = nx, ny
		px := int((x - minX) / (maxX - minX) * float64(c.W-1))
		py := int(float64(c.H-1) - (y-minY)/(maxY-minY)*float64(c.H-1))
		if i > 20 {
			c.Add(px, py, 0.5)
		}
	}
	c.Normalize()
}

// phyllotaxis arranges dots at the golden angle — Vogel's sunflower model.
func phyllotaxis(c *render.Canvas, q config.Quality) {
	cx, cy := float64(c.W)/2, float64(c.H)/2
	pa := c.PixelAspect
	maxR := math.Min(cx, cy*pa) * 0.95
	n := 1500 * (q.Samples() + 1)
	golden := math.Pi * (3 - math.Sqrt(5))
	for i := 0; i < n; i++ {
		theta := float64(i) * golden
		r := math.Sqrt(float64(i)/float64(n)) * maxR
		x := cx + r*math.Cos(theta)
		y := cy + r*math.Sin(theta)/pa
		ix, iy := int(x), int(y)
		c.Add(ix, iy, 0.6)
		c.Add(ix+1, iy, 0.15)
		c.Add(ix-1, iy, 0.15)
		c.Add(ix, iy+1, 0.15)
		c.Add(ix, iy-1, 0.15)
	}
}

// plasma is a sum-of-sines smooth gradient field — fractal-noise look.
func plasma(c *render.Canvas, q config.Quality) {
	pa := c.PixelAspect
	octaves := 3 + int(q)
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			fx := float64(x) / float64(c.W)
			fy := float64(y) / float64(c.H) * pa / 2
			v, amp, freq := 0.0, 1.0, 4.0
			for o := 0; o < octaves; o++ {
				v += amp * math.Sin(fx*freq*math.Pi+freq*0.7) * math.Cos(fy*freq*math.Pi+freq*1.3)
				v += amp * 0.5 * math.Sin((fx+fy)*freq*math.Pi*1.3)
				amp *= 0.55
				freq *= 1.9
			}
			v = (v + 2) / 4
			c.Set(x, y, v)
		}
	}
	c.Normalize()
}

// voronoi shades by edge proximity (F2 - F1 distance) to N random seeds.
func voronoi(c *render.Canvas, q config.Quality) {
	nSeeds := 12 + 8*int(q)
	r := rand.New(rand.NewSource(21))
	seeds := make([][2]float64, nSeeds)
	for i := range seeds {
		seeds[i] = [2]float64{r.Float64() * float64(c.W), r.Float64() * float64(c.H)}
	}
	pa := c.PixelAspect
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			fx, fy := float64(x), float64(y)
			d1, d2 := math.Inf(1), math.Inf(1)
			for _, s := range seeds {
				dx := fx - s[0]
				dy := (fy - s[1]) * pa
				d := math.Sqrt(dx*dx + dy*dy)
				if d < d1 {
					d2, d1 = d1, d
				} else if d < d2 {
					d2 = d
				}
			}
			edge := (d2 - d1) / (d2 + d1 + 1e-9)
			v := 1 - math.Min(1, edge*4)
			c.Set(x, y, v)
		}
	}
}

// truchet draws random arc tiles — endlessly varied but tessellating pattern.
func truchet(c *render.Canvas, q config.Quality) {
	tile := 8 + 4*int(q)
	if tile < 6 {
		tile = 6
	}
	r := rand.New(rand.NewSource(13))
	radius := float64(tile) / 2
	for ty := 0; ty < c.H; ty += tile {
		for tx := 0; tx < c.W; tx += tile {
			orient := r.Intn(2) == 0
			for dy := 0; dy < tile && ty+dy < c.H; dy++ {
				for dx := 0; dx < tile && tx+dx < c.W; dx++ {
					fdx, fdy := float64(dx), float64(dy)
					fr := float64(tile)
					var d1, d2 float64
					if orient {
						d1 = math.Sqrt(fdx*fdx + fdy*fdy)
						d2 = math.Sqrt((fr-fdx)*(fr-fdx) + (fr-fdy)*(fr-fdy))
					} else {
						d1 = math.Sqrt((fr-fdx)*(fr-fdx) + fdy*fdy)
						d2 = math.Sqrt(fdx*fdx + (fr-fdy)*(fr-fdy))
					}
					line := math.Min(math.Abs(d1-radius), math.Abs(d2-radius))
					v := 1 - line*0.6
					if v < 0 {
						v = 0
					}
					c.Set(tx+dx, ty+dy, v)
				}
			}
		}
	}
}
