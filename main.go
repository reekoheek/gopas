package main

/**
 * Imports
 */
import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/urfave/cli.v1"
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
			//Watches:    []string{"."},
			//Ignores:    []string{".git", ".gopath"},
			//Extensions: []string{"go"},
		},
	}

	tool.Bootstrap()

	app := cli.NewApp()
	app.Name = "gopas"
	app.Usage = "Go build tool outside GOPATH"
	app.Version = "0.1.0"

	app.Commands = []cli.Command{
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
				cli.StringFlag{
					Name: "exec, x",
				},
			},
			Action: tool.DoRun,
		},
		{
			Name:   "test",
			Usage:  "test project",
			Action: tool.DoTest,
		},
	}

	if err := app.Run(os.Args); err != nil {
		if !strings.HasPrefix(err.Error(), "exit status") {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		os.Exit(1)
	}
}
