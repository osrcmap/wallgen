package cmd

import "github.com/spf13/cobra"

func Root() *cobra.Command {
	root := &cobra.Command{
		Use:   "wallgen",
		Short: "ASCII / unicode wallpaper generator",
		Long: `Wallgen generates terminal wallpapers from math (fractals, curves)
or images. Interactive TUI asks for aspect ratio, quality, resolution,
and render style.`,
	}
	root.AddCommand(mathCmd())
	root.AddCommand(imageCmd())
	return root
}
