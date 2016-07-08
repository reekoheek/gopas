package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	GOPASFILE = "gopasfile"
)

/**
 * Dependency type
 */
type Dependency struct {
	Name, Version string
}

type Project interface {
	Bootstrap() error
	Dependencies() []Dependency
	Clean() error
	Install(dependency Dependency) error
	RunAsync(args []string) (*Runner, error)
	TestAsync() (*Runner, error)
	BuildAsync() (*Runner, error)

	Name() string
}

type ProjectImpl struct {
	Cwd string
	//Watches      []string
	//Ignores      []string
	//Extensions   []string
	dependencies []Dependency
	exeGo        string
}

func (p *ProjectImpl) Name() string {
	return filepath.Base(p.Cwd)
}

func (p *ProjectImpl) Dependencies() []Dependency {
	if p.dependencies == nil {
		p.dependencies = []Dependency{}
		if fileBytes, err := ioutil.ReadFile(filepath.Join(p.Cwd, GOPASFILE)); err == nil {
			fileLines := strings.Split(string(fileBytes), "\n")
			for _, line := range fileLines {
				if line != "" {
					token := strings.Split(line, "=")
					name := strings.Trim(token[0], " \t")
					version := ""
					if len(token) > 1 {
						version = strings.Trim(token[1], " \t")
					}
					p.dependencies = append(p.dependencies, Dependency{
						Name:    name,
						Version: version,
					})
				}
			}
		}
	}
	return p.dependencies
}

func (p *ProjectImpl) Install(dependency Dependency) error {
	gopathDir := filepath.Join(p.Cwd, ".gopath")
	srcDir := filepath.Join(gopathDir, "src")

	runner := &Runner{
		Name: p.exeGo,
		Args: []string{"get", dependency.Name},
		Dir:  srcDir,
		Env: []string{
			"GOPATH=" + gopathDir,
		},
	}
	err := runner.Run()
	if err != nil {
		return err
	}
	return runner.Wait()
}

func (p *ProjectImpl) BuildAsync() (*Runner, error) {
	gopathDir := filepath.Join(p.Cwd, ".gopath")
	projectDir := filepath.Join(gopathDir, "src", filepath.Base(p.Cwd))
	runner := &Runner{
		Name: p.exeGo,
		Args: []string{"build"},
		Dir:  projectDir,
		Env: []string{
			"GOPATH=" + gopathDir,
		},
	}
	err := runner.Run()
	return runner, err
}

func (p *ProjectImpl) RunAsync(args []string) (*Runner, error) {
	projectExe := filepath.Join(p.Cwd, filepath.Base(p.Cwd))
	runner := &Runner{
		Name: projectExe,
		Args: args,
	}

	err := runner.Run()
	return runner, err
}

func (p *ProjectImpl) TestAsync() (*Runner, error) {
	gopathDir := filepath.Join(p.Cwd, ".gopath")
	projectDir := filepath.Join(gopathDir, "src", filepath.Base(p.Cwd))
	runner := &Runner{
		Name: p.exeGo,
		Args: []string{"test"},
		Dir:  projectDir,
		Env: []string{
			"GOPATH=" + gopathDir,
		},
	}
	err := runner.Run()
	return runner, err
}

/**
 * Bootstrap project options and gopath dir
 */
func (p *ProjectImpl) Bootstrap() error {
	var err error
	if p.exeGo, err = exec.LookPath("go"); err != nil {
		return errors.New("Please install go")
	}

	if "" == p.Cwd {
		return errors.New("Cwd is undefined")
	}

	newCwd, err := filepath.Abs(p.Cwd)
	if err != nil {
		panic(err.Error())
	}
	p.Cwd = newCwd

	vendorDir := filepath.Join(p.Cwd, "vendor")
	gopathDir := filepath.Join(p.Cwd, ".gopath")
	gopathSrcDir := filepath.Join(gopathDir, "src")
	gopathProjectDir := filepath.Join(gopathSrcDir, filepath.Base(p.Cwd))

	if _, err := os.Stat(gopathDir); os.IsNotExist(err) {
		if os.MkdirAll(gopathSrcDir, 0755) != nil {
			panic(err.Error())
		}
	}

	// if err := os.RemoveAll(p.Cwd + "/.gopath/src/" + filepath.Base(p.Cwd)); err != nil {
	//  fmt.Fprintf(os.Stderr, ">> Bootstrap Error: %s\n", err.Error())
	// }
	// copy_folder(p.Cwd, p.Cwd+"/.gopath/src/"+filepath.Base(p.Cwd))

	os.Remove(gopathProjectDir)
	if err := os.Symlink(p.Cwd, gopathProjectDir); err != nil {
		panic(err.Error())
	}
	// FIXME workaround, cannot read module on vendor dir
	if files, err := ioutil.ReadDir(vendorDir); err == nil {
		for _, file := range files {
			name := file.Name()
			dest := filepath.Join(gopathSrcDir, name)

			os.Remove(dest)
			os.Symlink(filepath.Join(vendorDir, name), dest)
		}
	}

	return nil
}

func (p *ProjectImpl) Clean() error {
	os.RemoveAll(filepath.Join(p.Cwd, ".gopath"))
	os.Remove(filepath.Base(p.Cwd))
	return nil
}
