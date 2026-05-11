package render

import (
	"strings"

	"wallgen/internal/config"
)

// Ramps from dark to light.
var (
	asciiRamp  = []rune(" .,:;i1tfLCG08@")
	blockRamp  = []rune(" ░▒▓█")
	circleRamp = []rune(" ·∘○◍◉●")
)

// Render turns intensity canvas into a printable string.
func Render(c *Canvas, style config.Style) string {
	switch style {
	case config.StyleBraille:
		return renderBraille(c)
	case config.StyleBlock:
		return renderRamp(c, blockRamp)
	case config.StyleCircle:
		return renderRamp(c, circleRamp)
	default:
		return renderRamp(c, asciiRamp)
	}
}

func renderRamp(c *Canvas, ramp []rune) string {
	var b strings.Builder
	b.Grow((c.W + 1) * c.H)
	last := len(ramp) - 1
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			v := c.Get(x, y)
			idx := int(v * float64(last))
			if idx < 0 {
				idx = 0
			}
			if idx > last {
				idx = last
			}
			b.WriteRune(ramp[idx])
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// renderBraille packs 2x4 cells into a braille glyph for 8x density.
func renderBraille(c *Canvas) string {
	var b strings.Builder
	// braille base U+2800; bit map per dot position (col,row):
	//  (0,0)=0x01 (0,1)=0x02 (0,2)=0x04 (1,0)=0x08
	//  (1,1)=0x10 (1,2)=0x20 (0,3)=0x40 (1,3)=0x80
	dotMask := [2][4]uint16{
		{0x01, 0x02, 0x04, 0x40},
		{0x08, 0x10, 0x20, 0x80},
	}
	for y := 0; y < c.H; y += 4 {
		for x := 0; x < c.W; x += 2 {
			var glyph uint16 = 0x2800
			for dx := 0; dx < 2; dx++ {
				for dy := 0; dy < 4; dy++ {
					if c.Get(x+dx, y+dy) > 0.5 {
						glyph |= dotMask[dx][dy]
					}
				}
			}
			b.WriteRune(rune(glyph))
		}
		b.WriteByte('\n')
	}
	return b.String()
}
