package labee

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/LeBulldoge/labee/internal/database"
	"github.com/urfave/cli/v2"
)

type defaultAction func(ctx *cli.Context, args []string, db *database.DB) error

// Parses arguments from context and pipe, creates a DB connection,
// passes them on to the enclosed function.
func defaultActionWrapper(action defaultAction) cli.ActionFunc {
	return func(ctx *cli.Context) error {
		db, err := database.New()
		if err != nil {
			return fmt.Errorf("error initializing database connection: %w", err)
		}

		var args []string

		if ctx.Args().Present() {
			args = ctx.Args().Slice()
		}

		if pipeArgsAvailable() {
			args = append(args, readPipeArgs()...)
		}

		return action(ctx, args, db)
	}
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
					editLabel,
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
