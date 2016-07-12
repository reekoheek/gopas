package main

/**
 * Imports
 */
import (
	"fmt"
	"os"
	"strings"

	"github.com/reekoheek/gopas/util"

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

	logger := (&util.Logger{
		Out: os.Stdout,
		Err: os.Stderr,
	}).Construct()

	project, err := (&util.ProjectImpl{
		Cwd: cwd,
	}).Construct(logger)

	if err != nil {
		errHandler(err)
		return
	}

	tool, err := (&util.Tool{
		Project: project,
	}).Construct(logger)

	if err != nil {
		errHandler(err)
		return
	}

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
				Name:   "search",
				Usage:  "search from cache",
				Action: tool.DoSearch,
			},
			{
				Name:   "list",
				Usage:  "list dependencies",
				Action: tool.DoList,
			},
			{
				Name:   "run",
				Usage:  "run executable",
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
					&cli.StringFlag{
						Name:    "exec",
						Aliases: []string{"x"},
					},
				},
				Action: tool.DoWatch,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		errHandler(err)
		return
	}
}

func errHandler(err error) {
	if !strings.HasPrefix(err.Error(), "exit status") {
		fmt.Fprintf(os.Stderr, "--> E %s\n", err.Error())
	}
	os.Exit(1)
}
