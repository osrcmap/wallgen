package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"wallgen/internal/generator"
	"wallgen/internal/tui"
)

func imageCmd() *cobra.Command {
	choices := &tui.Choices{
		Aspect:     "16:9",
		Quality:    "med",
		Width:      "160",
		Style:      "ascii",
		Format:     "",
		Palette:    "mono",
		PixelScale: "6",
	}
	skipTUI := false
	cmd := &cobra.Command{
		Use:   "image [path]",
		Short: "Generate wallpaper from an image file",
		Long: `Convert a PNG/JPEG/GIF into a text or image wallpaper.

Output:
  text         (default) printable text wallpaper, optionally to .txt
  png / jpg    real image file (use --output and a palette)

Run without flags to use the interactive TUI. Pass a path as argument
or --path to populate the prompt.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				choices.ImagePath = args[0]
			}
			if !skipTUI {
				if err := tui.ImagePrompt(choices); err != nil {
					return err
				}
			}
			if choices.ImagePath == "" {
				return fmt.Errorf("image path required")
			}
			format := inferFormat(choices)
			if isPixelImageFormat(format) && choices.Width == "160" {
				choices.Width = "480"
			}
			w, h, style, q, pa, err := dims(choices, format)
			if err != nil {
				return err
			}
			canvas, err := generator.Image(choices.ImagePath, w, h, pa, q, choices.Invert)
			if err != nil {
				return err
			}
			if choices.OutputPath == "" && format == "text" {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"# %s | %s | %dx%d cells | quality=%s | invert=%v\n",
					choices.ImagePath, style, w, h, choices.Quality, choices.Invert)
			}
			return writeOutput(canvas, style, choices)
		},
	}
	f := cmd.Flags()
	f.StringVar(&choices.ImagePath, "path", "", "image path")
	f.StringVar(&choices.Aspect, "aspect", choices.Aspect, "aspect ratio (e.g. 16:9)")
	f.StringVar(&choices.Quality, "quality", choices.Quality, "low|med|high|ultra")
	f.StringVar(&choices.Width, "width", choices.Width, "width in cells (text) or pixels (png/jpg)")
	f.StringVar(&choices.Style, "style", choices.Style, "ascii|block|circle|braille (text only)")
	f.StringVar(&choices.Format, "format", "", "text|png|jpg (defaults to text, or inferred from -o ext)")
	f.StringVar(&choices.Palette, "palette", choices.Palette, "mono|inverse|fire|ocean|viridis|rainbow|sunset")
	f.StringVar(&choices.PixelScale, "scale", choices.PixelScale, "pixels per cell (png/jpg only)")
	f.BoolVar(&choices.Invert, "invert", false, "invert intensity (good for dark images)")
	f.StringVarP(&choices.OutputPath, "output", "o", "", "output file (default stdout for text)")
	f.BoolVar(&skipTUI, "no-tui", false, "skip interactive prompts, use flags only")
	return cmd
}
