package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/xueycn/ansisvg/ansitosvg"
	"github.com/xueycn/ansisvg/colorscheme/schemes"
)

type Env struct {
	Version  string
	ReadFile func(string) ([]byte, error)
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
	Args     []string
}

func Main(env Env) error {
	fs := flag.NewFlagSet("ansisvg", flag.ContinueOnError)
	var versionFlag bool
	fs.BoolVar(&versionFlag, "v", false, "")
	fs.BoolVar(&versionFlag, "version", false, "Show version")
	var fontNameFlag = fs.String("fontname", ansitosvg.DefaultOptions.FontName, "NAME|Font name")
	var fontFileFlag = fs.String("fontfile", "", "PATH|Font file to use and embed")
	var fontRefFlag = fs.String("fontref", "", "URL|External font URL to use")
	var fontSizeFlag = fs.Int("fontsize", ansitosvg.DefaultOptions.FontSize, "NUMBER|Font size")
	var terminalWidthFlag int
	fs.IntVar(&terminalWidthFlag, "w", 0, "")
	fs.IntVar(&terminalWidthFlag, "width", 0, "NUMBER|Terminal width (auto if not set)")
	var colorSchemeFlag = fs.String("colorscheme", ansitosvg.DefaultOptions.ColorScheme, "NAME|Color scheme")
	var listColorSchemesFlag = fs.Bool("listcolorschemes", false, "List color schemes")
	var transparentFlag = fs.Bool("transparent", ansitosvg.DefaultOptions.Transparent, "Transparent background")
	var gridModeFlag = fs.Bool("grid", false, "Grid mode (sets position for each character)")
	var helpFlag bool
	fs.BoolVar(&helpFlag, "h", false, "")
	fs.BoolVar(&helpFlag, "help", false, "Show help")
	charBoxSize := ansitosvg.DefaultOptions.CharBoxSize
	fs.Var(&charBoxSize, "charboxsize", "WxH|Character box size (use pixel units instead of font units)")
	marginSize := ansitosvg.DefaultOptions.MarginSize
	fs.Var(&marginSize, "marginsize", "WxH|Margin size (in either pixel or font units)")
	// handle error and usage output ourself
	fs.Usage = func() {}
	fs.SetOutput(io.Discard)
	longToShort := map[string]string{
		"help":    "h",
		"version": "v",
		"width":   "w",
	}
	flagHelp := func(f *flag.Flag) (string, string) {
		s := "--" + f.Name
		if short, ok := longToShort[f.Name]; ok {
			s += ", -" + short
		}

		arg, usage, found := strings.Cut(f.Usage, "|")
		if found {
			s += " " + arg
		} else {
			usage = arg
		}
		return s, usage
	}
	usage := func() {
		maxNameLen := 0
		fs.VisitAll(func(f *flag.Flag) {
			fh, _ := flagHelp(f)
			if len(fh) > maxNameLen {
				maxNameLen = len(fh)
			}
		})

		fmt.Fprintf(env.Stdout, `
%[1]s - Convert ANSI to SVG
Usage: %[1]s [FLAGS]

Example usage:
  program | %[1]s > file.svg

`[1:], fs.Name())
		fs.VisitAll(func(f *flag.Flag) {
			if len(f.Name) == 1 {
				return
			}

			fh, usage := flagHelp(f)
			pad := strings.Repeat(" ", maxNameLen-len(fh))
			fmt.Fprintf(env.Stdout, "%s%s  %s\n", fh, pad, usage)
		})
	}
	if err := fs.Parse(env.Args[1:]); err != nil {
		return err
	}
	if helpFlag {
		usage()
		return nil
	}

	if versionFlag {
		fmt.Fprintln(env.Stdout, env.Version)
		return nil
	}

	if *listColorSchemesFlag {
		maxNameLen := 0
		for _, n := range schemes.Names() {
			if len(n) > maxNameLen {
				maxNameLen = len(n)
			}
		}
		for _, n := range schemes.Names() {
			s, _ := schemes.Load(n)
			pad := strings.Repeat(" ", maxNameLen+1-len(n))
			fmt.Fprintf(env.Stdout, "%s\n", s.ANSIDemo(n+pad))
		}
		return nil
	}

	var fontEmbedded []byte
	if *fontFileFlag != "" {
		var err error
		fontEmbedded, err = env.ReadFile(*fontFileFlag)
		if err != nil {
			return err
		}
	}

	return ansitosvg.Convert(
		env.Stdin,
		env.Stdout,
		ansitosvg.Options{
			FontName:      *fontNameFlag,
			FontEmbedded:  fontEmbedded,
			FontRef:       *fontRefFlag,
			FontSize:      *fontSizeFlag,
			TerminalWidth: terminalWidthFlag,
			CharBoxSize:   charBoxSize,
			MarginSize:    marginSize,
			ColorScheme:   *colorSchemeFlag,
			Transparent:   *transparentFlag,
			GridMode:      *gridModeFlag,
		},
	)
}
