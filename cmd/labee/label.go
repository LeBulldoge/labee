package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/LeBulldoge/labee/internal/database"
	"github.com/gookit/color"
	"github.com/urfave/cli/v2"
)

func openInteractiveLabelMode(labels []database.Label) error {
	strs := []string{}
	for _, t := range labels {
		strs = append(strs, color.HEX(t.Color.String).Sprint(t.Name))
	}

	return openInteractiveMode(strs, "labee q l {}")
}

const (
	ColorNone = "none"
)

func isValidColor(hexColor string) (bool, error) {
	return regexp.MatchString("^#[0-9A-F]{6}$", hexColor)
}

func colorize(str string, hexColor string) (string, error) {
	if hexColor == ColorNone || hexColor == "" {
		return str, nil
	}

	validColor, err := isValidColor(hexColor)
	if err != nil {
		return "", err
	}

	if !validColor {
		return "", fmt.Errorf("%+v is not a valid HEX value", hexColor)
	}

	return color.HEX(hexColor).Sprint(str), nil
}

var (
	QueryLabel = &cli.Command{
		Name:    "label",
		Usage:   "Find files by their labels",
		Aliases: []string{"l"},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Show all available labels",
			},
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "Open an interactive view via fzf",
			},
		},
		Action: func(ctx *cli.Context) error {
			db, err := database.New()
			if err != nil {
				return err
			}

			args := []string{}
			if ctx.Args().Present() {
				args = ctx.Args().Slice()
			} else if ctx.Bool("all") {
				labels, err := db.GetAllLabels()
				if err != nil {
					return err
				}

				if len(labels) == 0 {
					return errors.New("no files found")
				}

				if ctx.Bool("interactive") {
					return openInteractiveLabelMode(labels)
				}

				cLabels := []string{}
				for _, t := range labels {
					cl, err := colorize(t.Name, t.Color.String)
					if err != nil {
						return err
					}

					cLabels = append(cLabels, cl)
				}

				color.Println(strings.Join(cLabels, ", "))

				return nil
			} else if pipeArgsAvailable() {
				args = append(args, readPipeArgs()...)
			} else {
				return ErrNoArgs
			}

			for _, v := range args {
				if !db.LabelExists(v) {
					e := fmt.Errorf("label '%s' does not exist", v)
					similar := db.GetSimilarLabel(v)
					if similar != nil {
						cl, _ := colorize(similar.Name, similar.Color.String)
						e = fmt.Errorf("%w. did you mean '%s'?", e, cl)
					}
					err = errors.Join(err, e)
				}
			}
			if err != nil {
				return err
			}

			files, err := db.GetFilesByLabels(args)
			if err != nil {
				return errors.New("no files found")
			}

			if ctx.Bool("interactive") {
				return openInteractiveFileMode(files)
			}

			// Just print out the file paths
			for _, f := range files {
				fmt.Println(f.Path)
			}

			return nil
		},
	}

	AddLabel = &cli.Command{
		Name:      "label",
		Usage:     "Create or configure an existing label",
		ArgsUsage: "[name]",
		Aliases:   []string{"t"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "color",
				Aliases: []string{"c"},
				Usage:   "provide a HEX color for the label",
			},
		},
		Action: func(ctx *cli.Context) error {
			if !ctx.Args().Present() {
				return ErrNoArgs
			}

			name := ctx.Args().First()
			color := ctx.String("color")

			db, err := database.New()
			if err != nil {
				return err
			}

			label, err := db.AddLabel(name, color)
			if err != nil {
				return err
			}

			log.Printf("label '%+v' edited", label)

			return nil
		},
	}

	RemoveLabel = &cli.Command{
		Name:      "label",
		Usage:     "Remove a label from the storage",
		ArgsUsage: "[name]",
		Aliases:   []string{"l"},
		Action: func(ctx *cli.Context) error {
			if !ctx.Args().Present() {
				return ErrNoArgs
			}
			args := ctx.Args().Slice()

			db, err := database.New()
			if err != nil {
				return err
			}

			for i := 0; i < len(args); i++ {
				err := db.DeleteLabel(args[i])
				if err != nil {
					return err
				}
			}

			fmt.Println("File removed from storage")

			return nil
		},
	}

	EditLabel = &cli.Command{
		Name:      "label",
		Usage:     "Edit a label",
		ArgsUsage: "[label to edit]",
		Aliases:   []string{"l"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Change the name",
			},
			&cli.StringFlag{
				Name:    "color",
				Aliases: []string{"c"},
				Usage:   "Set the color. HEX value or 'none'",
			},
		},
		Action: func(ctx *cli.Context) error {
			if !ctx.Args().Present() {
				return ErrNoArgs
			}
			label := ctx.Args().First()

			if !ctx.IsSet("name") || !ctx.IsSet("color") {
				return errors.New("please provide new values (name or color)")
			}

			db, err := database.New()
			if err != nil {
				return err
			}

			err = db.UpdateLabel(label, ctx.String("name"), ctx.String("color"))
			if err != nil {
				return err
			}

			log.Printf("Label edited. New value: %s", label)

			return nil
		},
	}
)
