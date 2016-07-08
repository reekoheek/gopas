package main

/**
 * Imports
 */
import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/urfave/cli.v2"
)

/**
 * main function
 */
func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}

	tool := &Tool{
		Project: &ProjectImpl{
			Cwd: cwd,
		},
	}

	tool.Bootstrap()

	app := &cli.App{
		Name:    "gopas",
		Usage:   "Go build tool outside GOPATH",
		Version: "0.1.0",
		Commands: []*cli.Command{
			{
				Name:   "build",
				Usage:  "build project",
				Action: tool.DoBuild,
			},
			{
				Name:   "clean",
				Usage:  "clean gopath",
				Action: tool.DoClean,
			},
			{
				Name:   "install",
				Usage:  "install dependencies",
				Action: tool.DoInstall,
			},
			{
				Name:   "list",
				Usage:  "list dependencies",
				Action: tool.DoList,
			},
			{
				Name:  "run",
				Usage: "run executable",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "exec",
						Aliases: []string{"x"},
					},
				},
				Action: tool.DoRun,
			},
			{
				Name:   "test",
				Usage:  "test project",
				Action: tool.DoTest,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "cover",
						Aliases: []string{"c"},
					},
				},
			},
			{
				Name:  "watch",
				Usage: "watch action",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:    "watch",
						Aliases: []string{"w"},
						Value:   cli.NewStringSlice("."),
					},
					&cli.StringFlag{
						Name:    "ext",
						Aliases: []string{"e"},
						Value:   "go",
					},
					&cli.StringSliceFlag{
						Name:    "ignore",
						Aliases: []string{"i"},
						Value:   cli.NewStringSlice(".git", ".gopath"),
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:   "run",
						Action: tool.DoWatchRun,
					},
					{
						Name:   "build",
						Action: tool.DoWatchBuild,
					},
					{
						Name:   "test",
						Action: tool.DoWatchTest,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		if !strings.HasPrefix(err.Error(), "exit status") {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		os.Exit(1)
	}
}
