package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/urfave/cli.v2"
)

type Tool struct {
	*Logger
	Project Project
}

func (t *Tool) Construct(logger *Logger) (*Tool, error) {
	t.Logger = logger

	if t.Project == nil {
		return t, errors.New("Project is undefined")
	}

	return t, nil
}

func (t *Tool) DoList(c *cli.Context) error {
	deps := t.Project.Dependencies()
	t.LogI("Dependencies %s (%d)", t.Project.Name(), len(deps))
	for _, dep := range deps {
		t.LogI("%s %s\n", dep.Name, dep.Version)
	}
	return nil
}

func (t *Tool) DoClean(c *cli.Context) error {
	t.LogI("Cleaning %s ...", t.Project.Name())
	return t.Project.Clean()
}

func (t *Tool) DoInstall(c *cli.Context) error {
	t.LogI("Installing %s ...", t.Project.Name())
	dependencies := t.Project.Dependencies()
	for _, dep := range dependencies {
		t.LogI("%s@%s", dep.Name, dep.Version)
		if err := t.Project.Install(dep); err != nil {
			t.LogE("  => fail, %s", err.Error())
		} else {
			t.LogI("  => ok")
		}
	}
	return nil
}

func (t *Tool) DoRun(c *cli.Context) error {
	if err := t.DoBuild(c); err != nil {
		return err
	}

	t.LogI("Running %s ...\n", t.Project.Name())
	if c == nil {
		return t.Project.Run()
	} else {
		return t.Project.Run(c.Args().Slice()...)
	}
}

func (t *Tool) DoBuild(c *cli.Context) error {
	t.LogI("Pre Building %s ...\n", t.Project.Name())
	if err := t.Project.PreBuild(); err != nil {
		return err
	}

	t.LogI("Building %s ...\n", t.Project.Name())
	return t.Project.Build()
}

func (t *Tool) DoTest(c *cli.Context) error {
	t.LogI("Testing %s ...\n", t.Project.Name())
	cover := c.Bool("cover")
	if cover {
		cwd, _ := os.Getwd()
		t.LogI(
			"Coverage html: %s",
			filepath.Join(cwd, ".gopath", "src", t.Project.Name(), "cover.html"))
	}
	return t.Project.Test(cover)
}

func (t *Tool) DoWatch(c *cli.Context) error {
	var (
		exeName string
		exeArgs []string
	)

	t.LogI("Watching %s ...\n", t.Project.Name())
	watcher := &Watcher{
		Logger:     t.Logger,
		Watches:    c.StringSlice("watch"),
		Extensions: strings.Split(c.String("ext"), ","),
		Ignores:    c.StringSlice("ignore"),
	}

	exec := c.String("exec")
	if exec != "" {
		splitted := strings.Split(exec, " ")
		exeName = splitted[0]
		exeArgs = splitted[1:]
	} else {
		slice := c.Args().Slice()

		exeName = os.Args[0]
		if len(slice) == 0 {
			exeArgs = []string{"run"}
		} else {
			exeArgs = slice
		}
	}

	return watcher.Watch(func() (*Runner, error) {
		//log.Println(exeName, exeArgs)
		runner := &Runner{
			Name: exeName,
			Args: exeArgs,
		}
		return runner, runner.Run()
	})
}
