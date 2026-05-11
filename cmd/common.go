package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/osrcmap/wallgen/internal/config"
	"github.com/osrcmap/wallgen/internal/render"
	"github.com/osrcmap/wallgen/internal/tui"
)

// dims resolves choices into final canvas dimensions and pixel aspect.
//
//	text         → terminal-cell canvas (pixelAspect 2.0); braille expands ×2,×4
//	png  / jpg   → square-pixel canvas (pixelAspect 1.0)
//	png-text /
//	jpg-text     → terminal-cell canvas (pixelAspect 2.0), then rasterized to
//	               an image of font glyphs by writeOutput.
func dims(c *tui.Choices, format string) (w, h int, style config.Style, q config.Quality, pixelAspect float64, err error) {
	asp, err := config.ParseAspect(c.Aspect)
	if err != nil {
		return 0, 0, "", 0, 0, err
	}
	width, err := strconv.Atoi(c.Width)
	if err != nil || width <= 0 {
		return 0, 0, "", 0, 0, fmt.Errorf("bad width %q", c.Width)
	}
	style = config.Style(c.Style)
	if style == "" {
		style = config.StyleASCII
	}
	q = config.ParseQuality(c.Quality)

	if isPixelImageFormat(format) {
		ph := width * asp.H / asp.W
		if ph < 4 {
			ph = 4
		}
		return width, ph, style, q, 1.0, nil
	}
	cw, ch := config.FitResolution(width, asp)
	pixelAspect = 2.0
	if style == config.StyleBraille {
		cw *= 2
		ch *= 4
		pixelAspect = 1.0
	}
	return cw, ch, style, q, pixelAspect, nil
}

func isImageFormat(f string) bool {
	return isPixelImageFormat(f) || isGlyphImageFormat(f)
}

func isPixelImageFormat(f string) bool {
	switch strings.ToLower(f) {
	case "png", "jpg", "jpeg":
		return true
	}
	return false
}

func isGlyphImageFormat(f string) bool {
	switch strings.ToLower(f) {
	case "png-text", "jpg-text", "jpeg-text":
		return true
	}
	return false
}

// inferFormat picks format from --format flag, falling back to file extension,
// finally "text".
func inferFormat(c *tui.Choices) string {
	if c.Format != "" {
		return strings.ToLower(c.Format)
	}
	lower := strings.ToLower(c.OutputPath)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "png"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "jpg"
	}
	return "text"
}

func writeOutput(canvas *render.Canvas, style config.Style, c *tui.Choices) error {
	format := inferFormat(c)
	scale, err := strconv.Atoi(c.PixelScale)
	if err != nil || scale < 1 {
		scale = 6
	}
	pal := render.ParsePalette(c.Palette)

	if isImageFormat(format) {
		if c.OutputPath == "" {
			return fmt.Errorf("--output required for %s format", format)
		}
		f, err := os.Create(c.OutputPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if isGlyphImageFormat(format) {
			return render.EncodeGlyphImage(canvas, pal, style, scale, format, f)
		}
		return render.EncodeImage(canvas, pal, scale, format, f)
	}
	out := render.Render(canvas, style)
	if c.OutputPath == "" {
		_, err := os.Stdout.WriteString(out)
		return err
	}
	return os.WriteFile(c.OutputPath, []byte(out), 0o644)
}
