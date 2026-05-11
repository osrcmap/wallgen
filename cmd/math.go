package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"wallgen/internal/generator"
	"wallgen/internal/tui"
)

func mathCmd() *cobra.Command {
	choices := &tui.Choices{
		Aspect:     "16:9",
		Quality:    "med",
		Width:      "160",
		Style:      "ascii",
		MathKind:   string(generator.Mandelbrot),
		Format:     "",
		Palette:    "viridis",
		PixelScale: "6",
	}
	skipTUI := false
	cmd := &cobra.Command{
		Use:   "math",
		Short: "Generate wallpaper from math (fractals, curves)",
		Long: `Generate ASCII / unicode / image wallpapers from mathematical patterns.
Patterns: mandelbrot, julia, sierpinski, lissajous, spiral, interference, rose.

Output:
  text         (default) printable text wallpaper, optionally to .txt
  png / jpg    real image file (use --output and a palette)

Run without flags to use the interactive TUI.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !skipTUI {
				if err := tui.MathPrompt(choices); err != nil {
					return err
				}
			}
			format := inferFormat(choices)
			if isPixelImageFormat(format) && choices.Width == "160" {
				choices.Width = "480" // 480 canvas px × default scale 4 ≈ 1920 final
			}
			w, h, style, q, pa, err := dims(choices, format)
			if err != nil {
				return err
			}
			kind := generator.MathKind(choices.MathKind)
			canvas, err := generator.Math(kind, w, h, pa, q)
			if err != nil {
				return err
			}
			if choices.OutputPath == "" && format == "text" {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"# %s | %s | %dx%d cells | quality=%s\n",
					kind, style, w, h, choices.Quality)
			}
			return writeOutput(canvas, style, choices)
		},
	}
	f := cmd.Flags()
	f.StringVar(&choices.MathKind, "kind", choices.MathKind, "math pattern")
	f.StringVar(&choices.Aspect, "aspect", choices.Aspect, "aspect ratio (e.g. 16:9)")
	f.StringVar(&choices.Quality, "quality", choices.Quality, "low|med|high|ultra")
	f.StringVar(&choices.Width, "width", choices.Width, "width in cells (text) or pixels (png/jpg)")
	f.StringVar(&choices.Style, "style", choices.Style, "ascii|block|circle|braille (text only)")
	f.StringVar(&choices.Format, "format", "", "text|png|jpg (defaults to text, or inferred from -o ext)")
	f.StringVar(&choices.Palette, "palette", choices.Palette, "mono|inverse|fire|ocean|viridis|rainbow|sunset")
	f.StringVar(&choices.PixelScale, "scale", choices.PixelScale, "pixels per cell (png/jpg only)")
	f.StringVarP(&choices.OutputPath, "output", "o", "", "output file (default stdout for text)")
	f.BoolVar(&skipTUI, "no-tui", false, "skip interactive prompts, use flags only")
	return cmd
}

