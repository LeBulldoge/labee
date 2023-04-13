package labee

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

const (
	colorNone = "none"
)

func isValidColor(hexColor string) (bool, error) {
	return regexp.MatchString("^#[0-9A-F]{6}$", hexColor)
}

func colorize(str string, hexColor string) (string, error) {
	if hexColor == colorNone || hexColor == "" {
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
	queryLabel = &cli.Command{
		Name:      "label",
		Usage:     "Find files by their labels",
		ArgsUsage: "[LABEL...]",
		Aliases:   []string{"l"},
		Flags: []cli.Flag{
			flagInteractive,
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Show all available labels",
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

				if interactive {
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

			for _, arg := range args {
				if !db.LabelExists(arg) {
					e := fmt.Errorf("label '%s' does not exist", arg)
					similar := db.GetSimilarLabel(arg)
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

			if interactive {
				return openInteractiveFileMode(files)
			}

			// Just print out the file paths
			for _, f := range files {
				fmt.Println(f.Path)
			}

			return nil
		},
	}

	removeLabel = &cli.Command{
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

	editLabel = &cli.Command{
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
