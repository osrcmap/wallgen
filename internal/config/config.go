package config

import (
	"fmt"
	"strconv"
	"strings"
)

type Aspect struct {
	W, H int
}

var Aspects = map[string]Aspect{
	"16:9":  {16, 9},
	"21:9":  {21, 9},
	"4:3":   {4, 3},
	"1:1":   {1, 1},
	"3:2":   {3, 2},
	"9:16":  {9, 16},
}

func ParseAspect(s string) (Aspect, error) {
	if a, ok := Aspects[s]; ok {
		return a, nil
	}
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Aspect{}, fmt.Errorf("invalid aspect %q", s)
	}
	w, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || w <= 0 || h <= 0 {
		return Aspect{}, fmt.Errorf("invalid aspect %q", s)
	}
	return Aspect{w, h}, nil
}

type Quality int

const (
	QLow Quality = iota
	QMed
	QHigh
	QUltra
)

func ParseQuality(s string) Quality {
	switch strings.ToLower(s) {
	case "low":
		return QLow
	case "high":
		return QHigh
	case "ultra":
		return QUltra
	default:
		return QMed
	}
}

// Iterations scales fractal/sample density by quality.
func (q Quality) Iterations() int {
	switch q {
	case QLow:
		return 60
	case QMed:
		return 150
	case QHigh:
		return 400
	case QUltra:
		return 1000
	}
	return 150
}

// Samples returns per-cell supersample count for image generator.
func (q Quality) Samples() int {
	switch q {
	case QLow:
		return 1
	case QMed:
		return 2
	case QHigh:
		return 3
	case QUltra:
		return 4
	}
	return 2
}

type Style string

const (
	StyleASCII   Style = "ascii"
	StyleBraille Style = "braille"
	StyleBlock   Style = "block"
	StyleCircle  Style = "circle"
)

// FitResolution scales width/height to match aspect while honouring user width.
// Terminal cells are roughly 2:1 (taller than wide), so we correct.
func FitResolution(width int, a Aspect) (int, int) {
	if width <= 0 {
		width = 120
	}
	// pixel height = width * (h/w) / cellAspect, cellAspect ~ 2
	h := int(float64(width) * float64(a.H) / float64(a.W) / 2.0)
	if h < 4 {
		h = 4
	}
	return width, h
}
