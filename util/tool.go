package util

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/urfave/cli.v2"
)

type Tool struct {
	*Logger
	Project Project
}

func NewTool(logger *Logger, project Project) (*Tool, error) {
	if project == nil {
		return nil, errors.New("Project is undefined")
	}

	t := &Tool{
		Logger:  logger,
		Project: project,
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

//func (t *Tool) DoSearch(c *cli.Context) error {
//	query := ""
//	if c.Args().Len() > 0 {
//		query = c.Args().First()
//	}
//	t.LogI("Searching %s: %s ...", t.Project.Name(), query)
//	return nil
//}

func (t *Tool) DoInstall(c *cli.Context) error {
	t.LogI("Installing %s ...", t.Project.Name())

	dependencies := t.Project.Dependencies()
	for _, dep := range dependencies {
		t.LogI("  Getting %s@%s", dep.Name, dep.Version)
		if err := t.Project.Get(dep); err != nil {
			t.LogE("  ---> %s@%s fail", dep.Name, dep.Version)
		}
	}

	baseDir := "_vendor/src"
	err := filepath.Walk(baseDir, func(path string, fi os.FileInfo, err error) error {
		if fi != nil && fi.IsDir() && path != baseDir {
			if fi.Name() == ".git" {
				return filepath.SkipDir
			} else {
				//fmt.Println("{{", path)
				err := filepath.Walk(path, func(childPath string, fi os.FileInfo, err error) error {
					if fi.IsDir() {
						if path == childPath {
							return nil
						} else {
							return filepath.SkipDir
						}
						//} else if fi.Name() == ".git" {
						//	return filepath.SkipDir
						//}
					}

					//fmt.Println("  --", childPath, err)
					if filepath.Ext(childPath) == ".go" {
						//fmt.Println("  >>", childPath, filepath.Ext(childPath))
						return PackageFound
					}
					return nil
				})
				//fmt.Println("}}", err)
				if err == PackageFound {
					pkg := path[len(baseDir)+1:]
					t.LogI("  Installing %s", pkg)
					if err = t.Project.GoRun("install", pkg); err != nil {
						return err
					}
					return filepath.SkipDir
				}
			}
		}
		return nil
	})
	return err
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
	if err := t.DoInstall(c); err != nil {
		return err
	}

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
	args := []string{}
	if c != nil {
		args = c.Args().Slice()
	}
	return t.Project.Test(cover, args...)
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
		fmt.Println("")
		runner := &Runner{
			Name: exeName,
			Args: exeArgs,
		}
		return runner, runner.Run()
	})
}

func (t *Tool) DoGo(c *cli.Context) error {
	args := c.Args().Slice()
	return t.Project.GoRun(args...)
}
