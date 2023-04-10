package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/LeBulldoge/labee/internal/database"
	"github.com/gookit/color"
	"github.com/urfave/cli/v2"
)

func openInteractiveFileMode(files []database.File) error {
	strs := []string{}
	for _, f := range files {
		strs = append(strs, f.Path)
	}
	return openInteractiveMode(strs, "labee q f {}")
}

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

var (
	QueryFile = &cli.Command{
		Name:      "file",
		Usage:     "Query by file",
		Aliases:   []string{"f"},
		ArgsUsage: "[keywords]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Show all matches or, if no arguments are provided, all files in the storage",
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
				return nil
			}

			args := ctx.Args().Slice()

			if ctx.Bool("all") {
				files, err := db.GetFiles(args)
				if err != nil {
					return err
				}

				if len(files) == 0 {
					return fmt.Errorf("no files found in storage")
				}

				if ctx.Bool("interactive") {
					return openInteractiveFileMode(files)
				}

				for _, f := range files {
					fmt.Println(f.Path)
				}

				return nil
			}

			if pipeArgsAvailable() {
				args = append(args, readPipeArgs()...)
			}

			if len(args) == 0 {
				return ErrNoArgs
			}

			files, err := db.GetFiles(args)
			if err != nil {
				return err
			}

			if ctx.Bool("interactive") {
				return openInteractiveFileMode(files)
			}

			file := files[0].Path

			labels, err := db.GetFileLabels(file)
			if err != nil {
				return fmt.Errorf("%s not found in the storage: %w", file, err)
			}

			printFileInfo(file, labels)

			return nil
		},
	}

	quiet   = false
	AddFile = &cli.Command{
		Name:      "file",
		Usage:     "Add a new file into the storage or add labels to existing files",
		ArgsUsage: "[path to file]",
		Aliases:   []string{"f"},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "quiet",
				Aliases:     []string{"q"},
				Usage:       "Mute output",
				Destination: &quiet,
			},
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
	}

	RemoveFile = &cli.Command{
		Name:      "file",
		Usage:     "Remove file(s) from the storage",
		ArgsUsage: "[absolute paths]",
		Aliases:   []string{"f"},
		Action: func(ctx *cli.Context) error {
			if ctx.Args().Present() {
				return ErrNoArgs
			}

			db, err := database.New()
			if err != nil {
				return err
			}

			args := ctx.Args().Slice()

			var errs error
			for i := 0; i < len(args); i++ {
				err := db.DeleteFile_b(args[i])
				if err != nil {
					errs = errors.Join(errs, err)
				}
			}

			if errs != nil {
				return errs
			}

			log.Println("File removed")

			return nil
		},
	}
)
