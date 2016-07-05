package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/urfave/cli.v1"
)

type Tool struct {
	Project Project
	Out     io.Writer
	Err     io.Writer
}

func (t *Tool) Bootstrap() error {
	if t.Project == nil {
		return errors.New("Project is undefined")
	}

	if t.Out == nil {
		t.Out = os.Stdout
	}

	if t.Err == nil {
		t.Err = os.Stderr
	}

	if err := t.Project.Bootstrap(); err != nil {
		return err
	}
	return nil
}

func (t *Tool) DoList(c *cli.Context) error {
	deps := t.Project.Dependencies()
	for _, dep := range deps {
		fmt.Fprintf(t.Out, "%s %s\n", dep.Name, dep.Version)
	}
	fmt.Fprintf(t.Out, "dependencies(%d)\n", len(deps))
	return nil
}

func (t *Tool) DoClean(c *cli.Context) error {
	fmt.Fprintln(t.Out, "Cleaning...")
	return t.Project.Clean()
}

func (t *Tool) DoInstall(c *cli.Context) error {
	fmt.Fprintln(t.Out, "Installing...")
	dependencies := t.Project.Dependencies()
	for _, dep := range dependencies {
		fmt.Fprintf(t.Out, "%s@%s => ", dep.Name, dep.Version)
		if err := t.Project.Install(dep); err != nil {
			fmt.Fprintln(t.Err, "fail")
		} else {
			fmt.Fprintln(t.Out, "ok")
		}
	}
	return nil
}

func (t *Tool) DoRun(c *cli.Context) error {
	fmt.Fprintln(t.Out, "Running...")
	return t.Project.Run()
}

func (t *Tool) DoBuild(c *cli.Context) error {
	fmt.Fprintln(t.Out, "Building...")
	return t.Project.Build()
}

func (t *Tool) DoTest(c *cli.Context) error {
	fmt.Fprintln(t.Out, "Testing...")
	return t.Project.Test()
}

func (t *Tool) DoHelp(c *cli.Context) error {
	fmt.Fprintln(t.Out, "Gopas is a tool to build Go outside GOPATH")
	fmt.Fprintln(t.Out, "")
	fmt.Fprintln(t.Out, "Usage:")
	fmt.Fprintln(t.Out, "")
	fmt.Fprintln(t.Out, "  gopas <action> [<args...>]")
	fmt.Fprintln(t.Out, "")
	fmt.Fprintln(t.Out, "The actions are:")
	fmt.Fprintln(t.Out, "")
	fmt.Fprintln(t.Out, "  build    compile packages and dependencies")
	fmt.Fprintln(t.Out, "  help     show help")
	fmt.Fprintln(t.Out, "  install  compile and install packages and dependencies")
	fmt.Fprintln(t.Out, "  list     list dependencies")
	fmt.Fprintln(t.Out, "  run      compile and run Go program")
	fmt.Fprintln(t.Out, "  clean    clean local .gopath")
	fmt.Fprintln(t.Out, "  test     test packages")
	fmt.Fprintln(t.Out, "")
	return nil
}
