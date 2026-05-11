package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/osrcmap/wallgen/internal/config"
	"github.com/osrcmap/wallgen/internal/generator"
)

// ---- validators ----

func validateAspect(s string) error {
	_, err := config.ParseAspect(strings.TrimSpace(s))
	if err != nil {
		return fmt.Errorf("expected W:H (e.g. 16:9)")
	}
	return nil
}

func validatePositiveInt(s string) error {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return fmt.Errorf("expected positive integer")
	}
	return nil
}

type Choices struct {
	Aspect     string
	Quality    string
	Width      string
	Style      string
	MathKind   string
	ImagePath  string
	Invert     bool
	OutputPath string
	Format     string
	Palette    string
	PixelScale string
}

const customSentinel = "__custom__"

type sentinelErr string

func (e sentinelErr) Error() string { return string(e) }

const errEmpty sentinelErr = "required"

// selectOnly shows a fixed-options select with no custom-typing escape hatch.
// Use for enums where free text would be invalid (style, format, palette,
// quality, math kind).
func selectOnly(title, desc string, opts []huh.Option[string], val *string) error {
	sel := huh.NewSelect[string]().
		Title(title).
		Description(desc).
		Options(opts...).
		Value(val)
	return huh.NewForm(huh.NewGroup(sel)).Run()
}

// selectOrCustom shows a select with an extra "custom…" option that opens a
// follow-up Input. Use for free-form values (aspect, width, scale).
// validate is run on the typed value (nil = only enforce non-empty).
func selectOrCustom(title, desc string, opts []huh.Option[string], val *string, validate func(string) error) error {
	full := make([]huh.Option[string], 0, len(opts)+1)
	full = append(full, opts...)
	full = append(full, huh.NewOption("custom… (type your own)", customSentinel))

	sel := huh.NewSelect[string]().
		Title(title).
		Description(desc).
		Options(full...).
		Value(val)
	if err := huh.NewForm(huh.NewGroup(sel)).Run(); err != nil {
		return err
	}
	if *val == customSentinel {
		*val = ""
		check := validate
		if check == nil {
			check = func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errEmpty
				}
				return nil
			}
		}
		in := huh.NewInput().
			Title("Custom value — " + title).
			Validate(check).
			Value(val)
		if err := huh.NewForm(huh.NewGroup(in)).Run(); err != nil {
			return err
		}
	}
	return nil
}

// cmdEcho accumulates flags chosen so far and reprints the equivalent CLI
// command after every step.
type cmdEcho struct {
	subcmd string
	parts  []string // "--flag value" entries, already shell-quoted
}

func newCmdEcho(subcmd string) *cmdEcho { return &cmdEcho{subcmd: subcmd} }

func (e *cmdEcho) add(flag, value string) {
	e.parts = append(e.parts, fmt.Sprintf("%s %s", flag, shellQuote(value)))
	e.print()
}

func (e *cmdEcho) addBool(flag string, value bool) {
	if value {
		e.parts = append(e.parts, flag)
	}
	e.print()
}

func (e *cmdEcho) print() {
	fmt.Fprintf(os.Stderr, "\n  $ wallgen %s --no-tui %s\n\n", e.subcmd, strings.Join(e.parts, " "))
}

func shellQuote(s string) string {
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, " \t'\"\\$`*?|&;<>()[]{}#~") {
		return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
	}
	return s
}

// ---- field helpers ----

func aspectStep(c *Choices, e *cmdEcho) error {
	opts := []huh.Option[string]{
		huh.NewOption("16:9 (widescreen)", "16:9"),
		huh.NewOption("21:9 (ultrawide)", "21:9"),
		huh.NewOption("4:3 (classic)", "4:3"),
		huh.NewOption("3:2 (photo)", "3:2"),
		huh.NewOption("1:1 (square)", "1:1"),
		huh.NewOption("9:16 (portrait)", "9:16"),
	}
	if err := selectOrCustom("Aspect ratio", "format W:H, e.g. 5:4", opts, &c.Aspect, validateAspect); err != nil {
		return err
	}
	e.add("--aspect", c.Aspect)
	return nil
}

func qualityStep(c *Choices, e *cmdEcho) error {
	opts := []huh.Option[string]{
		huh.NewOption("low    — fast preview",        "low"),
		huh.NewOption("medium — balanced (default)",  "med"),
		huh.NewOption("high   — sharp, slower",       "high"),
		huh.NewOption("ultra  — slowest, max detail", "ultra"),
	}
	if err := selectOnly("Quality", "iterations / supersample density", opts, &c.Quality); err != nil {
		return err
	}
	e.add("--quality", c.Quality)
	return nil
}

func widthStep(c *Choices, e *cmdEcho) error {
	opts := []huh.Option[string]{
		huh.NewOption("80",   "80"),
		huh.NewOption("120",  "120"),
		huh.NewOption("160",  "160"),
		huh.NewOption("200",  "200"),
		huh.NewOption("280",  "280"),
		huh.NewOption("400",  "400"),
		huh.NewOption("480 (px, ~1080p with scale 4)", "480"),
		huh.NewOption("720 (px, ~1440p with scale 4)", "720"),
		huh.NewOption("960 (px, ~4K with scale 4)",    "960"),
	}
	if err := selectOrCustom("Resolution / width",
		"text: chars across · png/jpg: pixels of canvas (× --scale = final image)",
		opts, &c.Width, validatePositiveInt); err != nil {
		return err
	}
	e.add("--width", c.Width)
	return nil
}

func styleStep(c *Choices, e *cmdEcho) error {
	opts := []huh.Option[string]{
		huh.NewOption("ascii  ( .:i1tCG@ )",  string(config.StyleASCII)),
		huh.NewOption("block  ( ░▒▓█ )",      string(config.StyleBlock)),
		huh.NewOption("circle ( ·∘○◍● )",     string(config.StyleCircle)),
		huh.NewOption("braille (max density)",string(config.StyleBraille)),
	}
	if err := selectOnly("Render style (text mode)", "ignored for png/jpg", opts, &c.Style); err != nil {
		return err
	}
	e.add("--style", c.Style)
	return nil
}

func formatStep(c *Choices, e *cmdEcho) error {
	opts := []huh.Option[string]{
		huh.NewOption("text       — terminal / .txt", "text"),
		huh.NewOption("png        — color blocks via palette",          "png"),
		huh.NewOption("jpg        — color blocks via palette",          "jpg"),
		huh.NewOption("png-text   — rasterized chars (uses --style)",   "png-text"),
		huh.NewOption("jpg-text   — rasterized chars (uses --style)",   "jpg-text"),
	}
	if err := selectOnly("Output format", "", opts, &c.Format); err != nil {
		return err
	}
	e.add("--format", c.Format)
	return nil
}

// formatUsesStyle reports whether the chosen format honours the text-style
// (ascii/block/circle/braille). Pixel image formats ignore style.
func formatUsesStyle(format string) bool {
	switch strings.ToLower(format) {
	case "png", "jpg", "jpeg":
		return false
	}
	return true
}

func paletteStep(c *Choices, e *cmdEcho) error {
	opts := []huh.Option[string]{
		huh.NewOption("mono (white on black)",   "mono"),
		huh.NewOption("inverse (black on white)","inverse"),
		huh.NewOption("fire",    "fire"),
		huh.NewOption("ocean",   "ocean"),
		huh.NewOption("viridis", "viridis"),
		huh.NewOption("rainbow", "rainbow"),
		huh.NewOption("sunset",  "sunset"),
	}
	if err := selectOnly("Color palette", "only used for png/jpg", opts, &c.Palette); err != nil {
		return err
	}
	e.add("--palette", c.Palette)
	return nil
}

func scaleStep(c *Choices, e *cmdEcho) error {
	opts := []huh.Option[string]{
		huh.NewOption("2",  "2"),
		huh.NewOption("4",  "4"),
		huh.NewOption("6",  "6"),
		huh.NewOption("8",  "8"),
		huh.NewOption("12", "12"),
		huh.NewOption("16", "16"),
	}
	if err := selectOrCustom("Pixel scale (pixels per cell)",
		"only used for png/jpg — multiplies canvas to final image size",
		opts, &c.PixelScale, validatePositiveInt); err != nil {
		return err
	}
	e.add("--scale", c.PixelScale)
	return nil
}

func mathKindStep(c *Choices, e *cmdEcho) error {
	opts := make([]huh.Option[string], 0, len(generator.MathKinds))
	for _, k := range generator.MathKinds {
		opts = append(opts, huh.NewOption(string(k), string(k)))
	}
	if err := selectOnly("Math pattern", "", opts, &c.MathKind); err != nil {
		return err
	}
	e.add("--kind", c.MathKind)
	return nil
}

func imagePathStep(c *Choices, e *cmdEcho) error {
	in := huh.NewInput().
		Title("Image path").
		Value(&c.ImagePath).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errEmpty
			}
			return nil
		})
	if err := huh.NewForm(huh.NewGroup(in)).Run(); err != nil {
		return err
	}
	e.add("--path", c.ImagePath)
	return nil
}

func invertStep(c *Choices, e *cmdEcho) error {
	conf := huh.NewConfirm().Title("Invert intensity?").Value(&c.Invert)
	if err := huh.NewForm(huh.NewGroup(conf)).Run(); err != nil {
		return err
	}
	e.addBool("--invert", c.Invert)
	return nil
}

func outputStep(c *Choices, e *cmdEcho) error {
	in := huh.NewInput().
		Title("Output file").
		Description("empty = stdout for text; required for png/jpg").
		Value(&c.OutputPath)
	if err := huh.NewForm(huh.NewGroup(in)).Run(); err != nil {
		return err
	}
	if c.OutputPath != "" {
		e.add("-o", c.OutputPath)
	} else {
		e.print()
	}
	return nil
}

// MathPrompt drives the full math wizard, printing the equivalent CLI command
// after every step. style step is skipped when format ignores it.
func MathPrompt(c *Choices) error {
	e := newCmdEcho("math")
	e.print()
	if err := mathKindStep(c, e); err != nil {
		return err
	}
	if err := aspectStep(c, e); err != nil {
		return err
	}
	if err := qualityStep(c, e); err != nil {
		return err
	}
	if err := widthStep(c, e); err != nil {
		return err
	}
	if err := formatStep(c, e); err != nil {
		return err
	}
	if formatUsesStyle(c.Format) {
		if err := styleStep(c, e); err != nil {
			return err
		}
	}
	if err := paletteStep(c, e); err != nil {
		return err
	}
	if err := scaleStep(c, e); err != nil {
		return err
	}
	return outputStep(c, e)
}

// ImagePrompt drives the full image wizard, printing the equivalent CLI
// command after every step. style step is skipped when format ignores it.
func ImagePrompt(c *Choices) error {
	e := newCmdEcho("image")
	e.print()
	if err := imagePathStep(c, e); err != nil {
		return err
	}
	if err := aspectStep(c, e); err != nil {
		return err
	}
	if err := qualityStep(c, e); err != nil {
		return err
	}
	if err := widthStep(c, e); err != nil {
		return err
	}
	if err := formatStep(c, e); err != nil {
		return err
	}
	if formatUsesStyle(c.Format) {
		if err := styleStep(c, e); err != nil {
			return err
		}
	}
	if err := paletteStep(c, e); err != nil {
		return err
	}
	if err := scaleStep(c, e); err != nil {
		return err
	}
	if err := invertStep(c, e); err != nil {
		return err
	}
	return outputStep(c, e)
}
