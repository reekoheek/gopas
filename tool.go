package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/urfave/cli.v2"
)

type Tool struct {
	Project Project
	Out     io.Writer
	Err     io.Writer
	iLogger *log.Logger
	eLogger *log.Logger
}

func (t *Tool) LogI(format string, args ...interface{}) {
	t.iLogger.Printf(format, args...)
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

	t.iLogger = log.New(t.Out, "--> I ", log.Lmicroseconds)
	t.eLogger = log.New(t.Err, "--> E ", log.Lmicroseconds)

	if err := t.Project.Bootstrap(); err != nil {
		return err
	}
	return nil
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
	t.LogI("Cleaning %s ...\n", t.Project.Name())
	return t.Project.Clean()
}

func (t *Tool) DoInstall(c *cli.Context) error {
	t.LogI("Installing %s ...\n", t.Project.Name())
	dependencies := t.Project.Dependencies()
	for _, dep := range dependencies {
		t.LogI("%s@%s\n", dep.Name, dep.Version)
		if err := t.Project.Install(dep); err != nil {
			fmt.Fprintln(t.Err, "=> fail")
		} else {
			fmt.Fprintln(t.Out, "=> ok")
		}
	}
	return nil
}

func (t *Tool) runAsync(c *cli.Context) (*Runner, error) {
	var args []string
	if c != nil {
		args = c.Args().Slice()
	} else {
		args = []string{}
	}

	exec := c.String("exec")
	if exec == "" {
		if err := t.DoBuild(c); err != nil {
			return nil, err
		}

		t.LogI("Running %s ...\n", t.Project.Name())
		return t.Project.RunAsync(args)
	} else {
		t.LogI("Running %s ...\n", t.Project.Name())
		execArr := strings.Split(exec, " ")
		runner := &Runner{
			Name: execArr[0],
			Args: execArr[1:],
		}
		return runner, runner.Run()
	}
}

func (t *Tool) DoRun(c *cli.Context) error {
	runner, err := t.runAsync(c)
	if err != nil {
		return err
	}
	if runner != nil {
		return runner.Wait()
	} else {
		return nil
	}
}

func (t *Tool) buildAsync(c *cli.Context) (*Runner, error) {
	t.LogI("Building %s ...\n", t.Project.Name())
	return t.Project.BuildAsync()
}

func (t *Tool) DoBuild(c *cli.Context) error {
	runner, err := t.buildAsync(c)
	if err != nil {
		return err
	}
	if runner != nil {
		return runner.Wait()
	} else {
		return nil
	}
}

func (t *Tool) testAsync(c *cli.Context) (*Runner, error) {
	t.LogI("Testing %s ...\n", t.Project.Name())
	cover := c.Bool("cover")
	if cover {
		cwd, _ := os.Getwd()
		t.LogI(
			"Coverage html: %s",
			filepath.Join(cwd, ".gopath", "src", t.Project.Name(), "cover.html"))
	}
	return t.Project.TestAsync(cover)
}

func (t *Tool) DoTest(c *cli.Context) error {
	runner, err := t.testAsync(c)
	if err != nil {
		return err
	}
	if runner != nil {
		return runner.Wait()
	} else {
		return nil
	}
}

func (t *Tool) DoWatchRun(c *cli.Context) error {
	t.LogI("Watching %s %s ...\n", t.Project.Name(), c.Command.Name)
	watcher := &Watcher{
		Watches:    c.StringSlice("watch"),
		Extensions: strings.Split(c.String("ext"), ","),
		Ignores:    c.StringSlice("ignore"),
		LogI:       t.LogI,
	}

	return watcher.Watch(func() (*Runner, error) {
		return t.runAsync(c)
	})
}

func (t *Tool) DoWatchBuild(c *cli.Context) error {
	t.LogI("Watching %s %s ...\n", t.Project.Name(), c.Command.Name)
	watcher := &Watcher{
		Watches:    c.StringSlice("watch"),
		Extensions: strings.Split(c.String("ext"), ","),
		Ignores:    c.StringSlice("ignore"),
		LogI:       t.LogI,
	}

	return watcher.Watch(func() (*Runner, error) {
		return t.buildAsync(c)
	})
}

func (t *Tool) DoWatchTest(c *cli.Context) error {
	t.LogI("Watching %s %s ...\n", t.Project.Name(), c.Command.Name)
	watcher := &Watcher{
		Watches:    c.StringSlice("watch"),
		Extensions: strings.Split(c.String("ext"), ","),
		Ignores:    c.StringSlice("ignore"),
		LogI:       t.LogI,
	}

	return watcher.Watch(func() (*Runner, error) {
		return t.testAsync(c)
	})
}
