package labee

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/LeBulldoge/labee/internal/database"
	"github.com/gookit/color"
	"github.com/urfave/cli/v2"
)

func openInteractiveMode(strs []string, preview string) error {
	cmd := exec.Command("fzf", "--ansi", "--height", "40%", "--border", "--preview", preview)

	pipe, err := cmd.StdinPipe()
	if err != nil {
		color.Danger.Println(err)
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return errors.New("interactive mode requires fzf to be installed and located in $PATH")
	}

	for _, s := range strs {
		color.Fprintln(pipe, s)
	}

	err = cmd.Wait()
	if err != nil && err.Error() != "exit status 130" {
		color.Danger.Println(err)
	}

	return nil
}

func Run() {
	app := &cli.App{
		Name:                   "labee",
		Usage:                  "Buzz around your files using labels!",
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		Commands: []*cli.Command{
			{
				Name:      "query",
				Usage:     "Query the storage",
				ArgsUsage: "[subcommand]",
				Aliases:   []string{"q"},
				Subcommands: []*cli.Command{
					queryFile,
					queryLabel,
				},
			},
			{
				Name:      "add",
				Usage:     "Attach labels to files",
				ArgsUsage: "[subcommand]",
				Aliases:   []string{"a"},
				Flags: []cli.Flag{
					flagQuiet,
					&cli.StringSliceFlag{
						Name:    "labels",
						Aliases: []string{"l"},
						Usage:   "Add comma separated labels to the file [-l \"labelA, labelB\"]. Creates labels if they don't exist",
					},
				},
				Action: func(ctx *cli.Context) error {
					var filepaths []string

					if !ctx.Args().Present() {
						if !pipeArgsAvailable() {
							return ErrNoArgs
						}
						filepaths = readPipeArgs()
					} else {
						filepaths = ctx.Args().Slice()
					}

					db, err := database.New()
					if err != nil {
						return err
					}

					labelNames := ctx.StringSlice("labels")

					var absPaths []string
					for _, v := range filepaths {
						_, err := os.Stat(v)
						if err != nil {
							return fmt.Errorf("file %s does not exist", v)
						}

						path, err := filepath.Abs(v)
						if err != nil {
							return err
						}

						absPaths = append(absPaths, path)
					}

					err = db.AddFilesAndLinks(absPaths, labelNames)
					if err != nil {
						return err
					}

					if quiet {
						return nil
					}

					for _, path := range absPaths {
						labels, err := db.GetFileLabels(path)
						if err != nil {
							return err
						}
						fmt.Println("New file added:")
						printFileInfo(path, labels)
					}

					return nil
				},
			},
			{
				Name:      "remove",
				Usage:     "Remove files or labels from the storage",
				ArgsUsage: "[subcommand]",
				Aliases:   []string{"r"},
				Subcommands: []*cli.Command{
					removeFile,
					removeLabel,
				},
			},
			{
				Name:    "edit",
				Usage:   "Edit items in the storage",
				Aliases: []string{"e"},
				Subcommands: []*cli.Command{
					EditLabel,
				},
			},
		},
	}

	log.SetPrefix("labee: ")
	log.SetFlags(log.Lmsgprefix)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
