package labee

import (
	"fmt"
	"log"
	"path/filepath"
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
		cLabels = append(cLabels, color.HEX(t.Color).Sprint(t.Name))
	}

	color.Println(strings.Join(cLabels, ", ") + "\n")
}

var removeFile = &cli.Command{
	Name:      "file",
	Usage:     "Remove file(s) from the storage",
	ArgsUsage: "[absolute paths]",
	Aliases:   []string{"f"},
	Action:    removeFileAction,
}

func removeFileAction(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if pipeArgsAvailable() {
		args = append(args, readPipeArgs()...)
	}

	if len(args) == 0 {
		return ErrNoArgs
	}

	db, err := database.FromContext(ctx.Context)
	if err != nil {
		return err
	}

	var paths []string
	for _, arg := range args {
		path, err := filepath.Abs(arg)
		if err != nil {
			return err
		}
		paths = append(paths, path)
	}

	err = db.DeleteFiles(ctx.Context, paths)
	if err != nil {
		return err
	}

	log.Printf("files %v removed", paths)

	return nil
}
