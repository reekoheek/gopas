package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"gopkg.in/yaml.v2"
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
	Dependencies() []Dependency
	Clean() error
	Install(dependency Dependency) error
	Run(args ...string) error
	Test(cover bool) error
	PreBuild() error
	Build() error

	Name() string
	Dir() string
}

type ProjectImpl struct {
	*Logger
	Cwd          string
	preBuild     [][]string
	dependencies []Dependency
	exeGo        string
}

func (p *ProjectImpl) Gopath() string {
	return filepath.Join(p.Cwd, ".gopath")
}

func (p *ProjectImpl) Env() []string {
	return []string{
		"GOPATH=" + p.Gopath(),
		"GOBIN=" + filepath.Join(p.Gopath(), "bin"),
	}
}

func (p *ProjectImpl) Dir() string {
	return filepath.Join(p.Gopath(), "src", p.Name())
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
	return p.GoRun("get", dependency.Name)
}

func (p *ProjectImpl) PreBuild() error {
	for _, cmdArr := range p.preBuild {
		p.LogI("  %s", cmdArr)

		runner := &Runner{
			Name: cmdArr[0],
			Args: cmdArr[1:],
			Dir:  p.Dir(),
		}

		if err := runner.Run(); err != nil {
			return err
		}

		if err := runner.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func (p *ProjectImpl) Build() error {
	return p.GoRun("build")
}

func (p *ProjectImpl) GoRun(args ...string) error {
	runner := &Runner{
		Name: p.exeGo,
		Args: args,
		Dir:  p.Dir(),
		Env:  p.Env(),
	}

	if err := runner.Run(); err != nil {
		return err
	}

	return runner.Wait()
}

func (p *ProjectImpl) Run(args ...string) error {
	executable := filepath.Join(p.Dir(), p.Name())

	runner := &Runner{
		Name: executable,
		Args: args,
	}

	cSignal := make(chan os.Signal, 1)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		//log.Println("[SIGNAL] caught ", sig, os.Args)
		if err := runner.Kill(); err != nil {
			fmt.Fprintf(os.Stderr, "[SIGNAL] Error: %s\n", err.Error())
		}
		// no need to exit, because process will immediately exit when child process exit
		//os.Exit(0)
	}()

	if err := runner.Run(); err != nil {
		return err
	}

	return runner.Wait()
}

func (p *ProjectImpl) Test(cover bool) error {
	if cover {
		if err := p.GoRun("test", "-coverprofile", "cover.out"); err != nil {
			return err
		}
		return p.GoRun("tool", "cover", "-html", "cover.out", "-o", "cover.html")
	} else {
		return p.GoRun("test")
	}
}

/**
 * Construct project options and gopath dir
 */
func (p *ProjectImpl) Construct(logger *Logger) (*ProjectImpl, error) {
	var err error

	p.Logger = logger

	if "" == p.Cwd {
		return p, errors.New("Cwd is undefined")
	}

	newCwd, err := filepath.Abs(p.Cwd)
	if err != nil {
		panic(err.Error())
	}
	p.Cwd = newCwd

	if p.exeGo, err = exec.LookPath("go"); err != nil {
		return p, errors.New("Please install go")
	}

	srcDir := filepath.Join(p.Gopath(), "src")
	if _, err = os.Stat(srcDir); os.IsNotExist(err) {
		if os.MkdirAll(srcDir, 0755) != nil {
			return p, err
		}
	}

	if err = os.RemoveAll(p.Dir()); err != nil {
		return p, err
	}

	if err = copy_folder(p.Cwd, p.Dir()); err != nil {
		return p, err
	}

	//os.Remove(gopathProjectDir)
	//if err := os.Symlink(p.Cwd, gopathProjectDir); err != nil {
	//	panic(err.Error())
	//}
	//// FIXME workaround, cannot read module on vendor dir
	//if files, err := ioutil.ReadDir(vendorDir); err == nil {
	//	for _, file := range files {
	//		name := file.Name()
	//		dest := filepath.Join(gopathSrcDir, name)

	//		os.Remove(dest)
	//		os.Symlink(filepath.Join(vendorDir, name), dest)
	//	}
	//}

	if content, err := ioutil.ReadFile(filepath.Join(p.Cwd, "gopas.yml")); err == nil {
		config := struct {
			PreBuild     [][]string `yaml:"pre-build"`
			Dependencies []string
		}{}
		if err = yaml.Unmarshal(content, &config); err == nil {
			p.preBuild = config.PreBuild
			for _, dep := range config.Dependencies {
				depSplitted := strings.Split(dep, "=")
				name := depSplitted[0]
				version := ""
				if len(depSplitted) > 1 {
					version = depSplitted[1]
				}
				p.dependencies = append(p.dependencies, Dependency{
					Name:    name,
					Version: version,
				})
			}
		}
	}

	return p, nil
}

func (p *ProjectImpl) Clean() error {
	os.RemoveAll(p.Gopath())
	return nil
}
