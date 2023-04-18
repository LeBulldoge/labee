package labee

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/LeBulldoge/labee/internal/database"
	"github.com/urfave/cli/v2"
)

func beforeHook(ctx *cli.Context) error {
	db, err := database.New()
	if err != nil {
		return err
	}

	ctx.Context = database.WithDatabase(ctx.Context, db)
	return nil
}

func afterHook(ctx *cli.Context) error {
	db, err := database.FromContext(ctx.Context)
	if err != nil {
		return err
	}

	return db.Close()
}

func Run() {
	app := &cli.App{
		Name:                   "labee",
		Usage:                  "Buzz around your files using labels!",
		UseShortOptionHandling: true,
		EnableBashCompletion:   true,
		Before:                 beforeHook,
		After:                  afterHook,
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

					labelNames := ctx.StringSlice("labels")

					var absPaths []string
					for _, v := range args {
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
