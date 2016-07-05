package main

import (
	"errors"
	"io/ioutil"
	"os"
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
	Run() error
	Test() error
	Build() error
}

type ProjectImpl struct {
	Cwd          string
	Watches      []string
	Ignores      []string
	Extensions   []string
	dependencies []Dependency
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
		Name: "go",
		Args: []string{"get", dependency.Name},
		Dir:  srcDir,
		Env: []string{
			"GOPATH=" + gopathDir,
		},
	}
	command, err := runner.Run()
	if err != nil {
		return err
	}
	command.Wait()
	return nil
}

func (p *ProjectImpl) Build() error {
	gopathDir := filepath.Join(p.Cwd, ".gopath")
	projectDir := filepath.Join(gopathDir, "src", filepath.Base(p.Cwd))
	runner := &Runner{
		Name: "go",
		Args: []string{"build"},
		Dir:  projectDir,
		Env: []string{
			"GOPATH=" + gopathDir,
		},
	}
	command, err := runner.Run()
	if err != nil {
		return err
	}
	command.Wait()
	return nil
}

func (p *ProjectImpl) Run() error {
	p.Build()

	gopathDir := filepath.Join(p.Cwd, ".gopath")
	projectBase := filepath.Base(p.Cwd)
	projectDir := filepath.Join(gopathDir, "src", projectBase)
	projectExe := filepath.Join(projectDir, projectBase)
	runner := &Runner{
		Name: projectExe,
	}
	command, err := runner.Run()
	if err != nil {
		return err
	}
	command.Wait()
	return nil
}

func (p *ProjectImpl) Test() error {
	gopathDir := filepath.Join(p.Cwd, ".gopath")
	projectDir := filepath.Join(gopathDir, "src", filepath.Base(p.Cwd))
	runner := &Runner{
		Name: "go",
		Args: []string{"test"},
		Dir:  projectDir,
		Env: []string{
			"GOPATH=" + gopathDir,
		},
	}
	command, err := runner.Run()
	if err != nil {
		return err
	}
	command.Wait()
	return nil
}

/**
 * Bootstrap project options and gopath dir
 */
func (p *ProjectImpl) Bootstrap() error {
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
	if err := os.RemoveAll(filepath.Join(p.Cwd, ".gopath")); err != nil {
		return err
	}
	if err := os.Remove(filepath.Base(p.Cwd)); err != nil {
		return err
	}
	return nil
}
