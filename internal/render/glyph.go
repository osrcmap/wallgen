package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"sync"
	"unicode/utf8"

	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/osrcmap/wallgen/internal/config"
)

var (
	glyphFontOnce sync.Once
	glyphFont     *opentype.Font
	glyphFontErr  error
)

func loadGlyphFont() (*opentype.Font, error) {
	glyphFontOnce.Do(func() {
		glyphFont, glyphFontErr = opentype.Parse(gomono.TTF)
	})
	return glyphFont, glyphFontErr
}

// EncodeGlyphImage rasterizes the canvas as a grid of font glyphs (chosen by
// `style`) tinted per-cell by `palette`. cellPx is the side of the smaller
// dimension of one text cell (cells render 2:1 tall × wide).
//
// braille is a special case: gomono lacks U+2800 glyphs, so we draw the 2×4
// sub-cell dots directly as filled rectangles.
func EncodeGlyphImage(c *Canvas, p Palette, style config.Style, cellPx int, format string, w io.Writer) error {
	if cellPx < 6 {
		cellPx = 6
	}
	if style == config.StyleBraille {
		return encodeBrailleDots(c, p, cellPx, format, w)
	}
	f, err := loadGlyphFont()
	if err != nil {
		return fmt.Errorf("load font: %w", err)
	}
	cw, ch := cellPx, cellPx*2

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    float64(ch),
		DPI:     72,
		Hinting: xfont.HintingFull,
	})
	if err != nil {
		return fmt.Errorf("font face: %w", err)
	}
	defer face.Close()

	text := Render(c, style)
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) == 0 {
		return fmt.Errorf("empty text render")
	}
	nrows := len(lines)
	ncols := 0
	for _, l := range lines {
		if rc := utf8.RuneCountInString(l); rc > ncols {
			ncols = rc
		}
	}

	pixW, pixH := ncols*cw, nrows*ch
	img := image.NewRGBA(image.Rect(0, 0, pixW, pixH))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)

	// Sub-cell sampling: braille packs 2×4 canvas cells per glyph; others 1×1.
	sx, sy := 1, 1
	if style == config.StyleBraille {
		sx, sy = 2, 4
	}

	for row, line := range lines {
		col := 0
		for _, r := range line {
			if !isBlank(r) {
				v := sampleRegion(c, col*sx, row*sy, sx, sy)
				d := xfont.Drawer{
					Dst:  img,
					Src:  image.NewUniform(p.Sample(v)),
					Face: face,
					Dot: fixed.Point26_6{
						X: fixed.I(col * cw),
						Y: fixed.I(row*ch + ch - cellPx/3),
					},
				}
				d.DrawString(string(r))
			}
			col++
		}
	}

	switch strings.ToLower(format) {
	case "jpg-text", "jpeg-text":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 92})
	case "png-text":
		return png.Encode(w, img)
	default:
		return fmt.Errorf("unsupported glyph format %q", format)
	}
}

func isBlank(r rune) bool {
	return r == ' ' || r == '⠀' || r == '\t'
}

// encodeBrailleDots paints each lit sub-cell of the (already 2×4-expanded)
// canvas as a filled circle, in palette colour. One braille glyph = 2×4 dots
// laid out as a (cellPx) × (2*cellPx) rectangle.
func encodeBrailleDots(c *Canvas, p Palette, cellPx int, format string, w io.Writer) error {
	cw, ch := cellPx, cellPx*2 // outer braille glyph cell
	subW, subH := cw/2, ch/4   // sub-cell footprint
	if subW < 1 {
		subW = 1
	}
	if subH < 1 {
		subH = 1
	}
	dotR := subW
	if subH < dotR {
		dotR = subH
	}
	dotR = dotR / 2
	if dotR < 1 {
		dotR = 1
	}
	gcols := c.W / 2
	grows := c.H / 4
	pixW, pixH := gcols*cw, grows*ch
	img := image.NewRGBA(image.Rect(0, 0, pixW, pixH))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)

	for grow := 0; grow < grows; grow++ {
		for gcol := 0; gcol < gcols; gcol++ {
			for dx := 0; dx < 2; dx++ {
				for dy := 0; dy < 4; dy++ {
					sx, sy := gcol*2+dx, grow*4+dy
					v := c.Get(sx, sy)
					if v <= 0.5 {
						continue
					}
					col := p.Sample(v)
					cx := gcol*cw + dx*subW + subW/2
					cy := grow*ch + dy*subH + subH/2
					fillDisk(img, cx, cy, dotR, col)
				}
			}
		}
	}

	switch strings.ToLower(format) {
	case "jpg-text", "jpeg-text":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 92})
	case "png-text":
		return png.Encode(w, img)
	default:
		return fmt.Errorf("unsupported glyph format %q", format)
	}
}

func fillDisk(img *image.RGBA, cx, cy, r int, col color.RGBA) {
	r2 := r * r
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r2 {
				img.SetRGBA(cx+dx, cy+dy, col)
			}
		}
	}
}

func sampleRegion(c *Canvas, x, y, w, h int) float64 {
	var sum float64
	var n int
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			ix, iy := x+dx, y+dy
			if ix >= 0 && iy >= 0 && ix < c.W && iy < c.H {
				sum += c.Get(ix, iy)
				n++
			}
		}
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}
