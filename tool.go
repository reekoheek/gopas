package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

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
	fmt.Fprintf(t.Out, "Dependencies %s (%d)\n", t.Project.Name(), len(deps))
	for _, dep := range deps {
		fmt.Fprintf(t.Out, "%s %s\n", dep.Name, dep.Version)
	}
	return nil
}

func (t *Tool) DoClean(c *cli.Context) error {
	fmt.Fprintf(t.Out, "Cleaning %s ...\n", t.Project.Name())
	return t.Project.Clean()
}

func (t *Tool) DoInstall(c *cli.Context) error {
	fmt.Fprintf(t.Out, "Installing %s ...\n", t.Project.Name())
	dependencies := t.Project.Dependencies()
	for _, dep := range dependencies {
		fmt.Fprintf(t.Out, "%s@%s\n", dep.Name, dep.Version)
		if err := t.Project.Install(dep); err != nil {
			fmt.Fprintln(t.Err, "=> fail")
		} else {
			fmt.Fprintln(t.Out, "=> ok")
		}
	}
	return nil
}

func (t *Tool) DoRun(c *cli.Context) error {
	exec := c.String("exec")

	if exec == "" {
		if err := t.DoBuild(c); err != nil {
			return err
		}

		fmt.Fprintf(t.Out, "Running %s ...\n", t.Project.Name())
		return t.Project.Run(c.Args())
	} else {
		fmt.Fprintf(t.Out, "Running %s ...\n", t.Project.Name())
		execArr := strings.Split(exec, " ")
		runner := &Runner{
			Name: execArr[0],
			Args: execArr[1:],
		}
		cmd, _ := runner.Run()
		return cmd.Wait()
	}
}

func (t *Tool) DoBuild(c *cli.Context) error {
	fmt.Fprintf(t.Out, "Building %s ...\n", t.Project.Name())
	return t.Project.Build()
}

func (t *Tool) DoTest(c *cli.Context) error {
	fmt.Fprintf(t.Out, "Testing %s ...\n", t.Project.Name())
	return t.Project.Test()
}
