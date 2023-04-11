package labee

import (
	"errors"
	"fmt"
	"strings"

	"github.com/LeBulldoge/labee/internal/database"
	"github.com/gookit/color"
	"github.com/urfave/cli/v2"
)

func printFileInfo(file string, labels []database.Label) {
	color.Tag("us").Println(file)

	if len(labels) == 0 {
		fmt.Println("No labels have been assigned")
		return
	}

	fmt.Print("Labels: ")

	cLabels := []string{}
	for _, t := range labels {
		cLabels = append(cLabels, color.HEX(t.Color.String).Sprint(t.Name))
	}

	color.Println(strings.Join(cLabels, ", ") + "\n")
}

var queryFile = &cli.Command{
	Name:      "file",
	Usage:     "Find a file in storage via keywords",
	Aliases:   []string{"f"},
	ArgsUsage: "[keywords]",
	Flags: []cli.Flag{
		flagInteractive,
		&cli.BoolFlag{
			Name:    "all",
			Aliases: []string{"a"},
			Usage:   "Show all matches or, if no arguments are provided, all files in the storage",
		},
	},
	Action: defaultActionWrapper(queryFileAction),
}

func queryFileAction(ctx *cli.Context, args []string, db *database.DB) error {
	if ctx.Bool("all") {
		files, err := db.GetFiles(args)
		if err != nil {
			return err
		}

		if interactive {
			return openInteractiveFileMode(files)
		}

		for _, f := range files {
			fmt.Println(f.Path)
		}

		return nil
	}

	if len(args) == 0 {
		return ErrNoArgs
	}

	files, err := db.GetFiles(args)
	if err != nil {
		return err
	}

	if interactive {
		return openInteractiveFileMode(files)
	}

	file := files[0].Path

	labels, err := db.GetFileLabels(file)
	if err != nil {
		return fmt.Errorf("could not find labels for file %s: %w", file, err)
	}

	printFileInfo(file, labels)

	return nil
}

var removeFile = &cli.Command{
	Name:      "file",
	Usage:     "Remove file(s) from the storage",
	ArgsUsage: "[absolute paths]",
	Aliases:   []string{"f"},
	Action:    defaultActionWrapper(removeFileAction),
}

func removeFileAction(_ *cli.Context, args []string, db *database.DB) error {
	if len(args) == 0 {
		return ErrNoArgs
	}

	var errs error
	for i := 0; i < len(args); i++ {
		err := db.DeleteFile_b(args[i])
		fmt.Printf("%s removed", args[i])
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if errs != nil {
		return errs
	}

	return nil
}
