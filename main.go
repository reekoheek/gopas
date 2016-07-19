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
				Name:    "build",
				Aliases: []string{"b"},
				Usage:   "build project",
				Action:  tool.DoBuild,
			},
			{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   "clean gopath",
				Action:  tool.DoClean,
			},
			{
				Name:    "install",
				Aliases: []string{"i"},
				Usage:   "install dependencies",
				Action:  tool.DoInstall,
			},
			{
				Name:    "search",
				Aliases: []string{"s"},
				Usage:   "search from cache",
				Action:  tool.DoSearch,
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list dependencies",
				Action:  tool.DoList,
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "run executable",
				Action:  tool.DoRun,
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "test project",
				Action:  tool.DoTest,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "cover",
						Aliases: []string{"c"},
					},
				},
			},
			{
				Name:    "go",
				Aliases: []string{"g"},
				Usage:   "invoke go command",
				Action:  tool.DoGo,
			},
			{
				Name:    "watch",
				Aliases: []string{"w"},
				Usage:   "watch action",
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
