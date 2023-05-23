package labee

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/LeBulldoge/labee/internal/database"
	"github.com/gookit/color"
	"github.com/urfave/cli/v2"
)

const (
	colorNone = "NONE"
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

func doLabelsExist(db *database.DB, labelNames []string) error {
	var err error
	for _, label := range labelNames {
		if db.LabelExists(label) {
			continue
		}

		e := fmt.Errorf("label '%s' does not exist", label)
		if len(label) >= 3 {
			if similar := db.GetSimilarLabel(label); similar != nil {
				cl, _ := colorize(similar.Name, similar.Color)
				e = fmt.Errorf("%w. did you mean '%s'?", e, cl)
			}
		}
		err = errors.Join(err, e)
	}

	return err
}

var (
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

			db, err := database.New(ctx.Context)
			if err != nil {
				return err
			}

			for i := 0; i < len(args); i++ {
				err := db.DeleteLabel(ctx.Context, args[i])
				if err != nil {
					return err
				}
			}

			fmt.Println("File removed from storage")

			return nil
		},
	}

	editLabel = &cli.Command{
		Name:      "edit",
		Usage:     "Change the name or color of a label",
		ArgsUsage: "[label to edit]",
		Aliases:   []string{"e"},
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

			if !ctx.IsSet("name") && !ctx.IsSet("color") {
				return errors.New("please provide new values (name or color)")
			}

			db, err := database.New(ctx.Context)
			if err != nil {
				return err
			}

			err = db.UpdateLabel(ctx.Context, label, ctx.String("name"), ctx.String("color"))
			if err != nil {
				return err
			}

			log.Printf("Label edited.")

			return nil
		},
	}
)
