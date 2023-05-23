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
				Name:      "find",
				Usage:     "Query the storage for files. Filter by labels, filename or location.",
				ArgsUsage: "[PATH]",
				Aliases:   []string{"f"},
				Flags: []cli.Flag{
					flagInteractive,
					&cli.StringSliceFlag{
						Name:    "labels",
						Aliases: []string{"l"},
						Usage:   "List of comma separated labels",
					},
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "Glob pattern to filter the filenames with",
					},
				},
				Action: func(ctx *cli.Context) error {
					db, err := database.FromContext(ctx.Context)
					if err != nil {
						return err
					}

					labels := ctx.StringSlice("labels")
					if err := doLabelsExist(db, labels); err != nil {
						return err
					}

					pattern := ctx.String("name")

					var pathPrefix string
					if ctx.Args().Present() {
						pathPrefix, err = filepath.Abs(ctx.Args().First())
						if err != nil {
							return err
						}
					}

					var files []database.File
					if len(labels) > 0 {
						files, err = db.GetFilesFilteredWithLabels(labels, pattern, pathPrefix)
						if err != nil {
							return err
						}
					} else {
						files, err = db.GetFilesFiltered(pattern, pathPrefix)
						if err != nil {
							return err
						}
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
			},
			{
				Name:      "info",
				Usage:     "Print out information about the specified files",
				ArgsUsage: "[PATH]",
				Aliases:   []string{"i"},
				Action: func(ctx *cli.Context) error {
					db, err := database.FromContext(ctx.Context)
					if err != nil {
						return err
					}

					if !ctx.Args().Present() {
						return ErrNoArgs
					}

					filenames := ctx.Args().Slice()
					for _, filename := range filenames {
						labels, err := db.GetFileLabels(filename)
						if err != nil {
							return err
						}

						printFileInfo(filename, labels)
					}

					return nil
				},
			},
			{
				Name:      "add",
				Usage:     "Attach labels to files",
				ArgsUsage: "[FILE...]",
				Aliases:   []string{"a"},
				Flags: []cli.Flag{
					flagQuiet,
					&cli.StringSliceFlag{
						Name:    "labels",
						Aliases: []string{"l"},
						Usage:   "Add comma separated labels to the file [-l \"labelA, labelB\"]. Creates labels if they don't exist",
					},
				},
				Action: addLink,
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
			editLabel,
		},
	}

	log.SetPrefix("labee: ")
	log.SetFlags(log.Lmsgprefix)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func addLink(ctx *cli.Context) error {
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
}
