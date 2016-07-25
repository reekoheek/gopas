package util

import (
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
type (
	Dependency struct {
		Name    string
		Version string
	}

	Project interface {
		Dependencies() []Dependency
		Clean() error
		Get(dependency Dependency) error
		Run(args ...string) error
		Test(cover bool, packages ...string) error
		PreBuild() error
		Build() error

		Name() string
		Dir() string
		GoRun(args ...string) error
	}

	ProjectImpl struct {
		*Logger
		Cwd          string
		name         string
		gopaths      []string
		preBuild     [][]string
		dependencies []Dependency
		exeGo        string
	}
)

func (p *ProjectImpl) Gopath() []string {
	if len(p.gopaths) == 0 {
		p.gopaths = []string{
			filepath.Join(p.Cwd, ".gopath"),
			filepath.Join(p.Cwd, "_vendor"),
		}
	}

	return p.gopaths
}

func (p *ProjectImpl) Env() []string {
	return []string{
		"GOPATH=" + strings.Join(p.Gopath(), ":"),
	}
}

func (p *ProjectImpl) Dir() string {
	return filepath.Join(p.Gopath()[0], "src", p.Name())
}

func (p *ProjectImpl) Name() string {
	if p.name == "" {
		p.name = filepath.Base(p.Cwd)
	}
	return p.name
}

func (p *ProjectImpl) Dependencies() []Dependency {
	p.Bootstrap()
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

func (p *ProjectImpl) Get(dependency Dependency) error {
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
	return p.GoRun("install")
}

func (p *ProjectImpl) GoRun(args ...string) error {
	if err := p.Bootstrap(); err != nil {
		return err
	}
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
	executable := filepath.Join(p.Gopath()[0], "bin", filepath.Base(p.Name()))

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

func (p *ProjectImpl) Test(cover bool, packages ...string) error {
	createArgs := func(args []string) []string {
		args = append(args, packages...)
		return args
	}

	if cover {
		args := createArgs([]string{"test", "-coverprofile", "cover.out"})
		if err := p.GoRun(args...); err != nil {
			return err
		}

		if _, err := os.Stat(filepath.Join(p.Dir(), "cover.out")); os.IsNotExist(err) {
			return nil
		}

		return p.GoRun("tool", "cover", "-html", "cover.out", "-o", "cover.html")
	} else {
		args := createArgs([]string{"test"})
		return p.GoRun(args...)
	}
}

func (p *ProjectImpl) Clean() error {
	os.RemoveAll(p.Gopath()[0])
	return nil
}

func (p *ProjectImpl) Bootstrap() error {
	var (
		err     error
		content []byte
	)

	if p.exeGo != "" {
		return nil
	}
	if p.exeGo, err = exec.LookPath("go"); err != nil {
		panic("Please install go")
	}

	srcDir := filepath.Join(p.Gopath()[0], "src")
	if _, err = os.Stat(srcDir); os.IsNotExist(err) {
		if os.MkdirAll(srcDir, 0755) != nil {
			return err
		}
	}

	if content, err = ioutil.ReadFile(filepath.Join(p.Cwd, "gopas.yml")); err == nil {
		config := struct {
			Name         string
			PreBuild     [][]string `yaml:"pre-build"`
			Dependencies []string
		}{}
		if err = yaml.Unmarshal(content, &config); err == nil {
			p.name = config.Name
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

	if err = os.RemoveAll(p.Dir()); err != nil {
		return err
	}

	if err = copy_folder(p.Cwd, p.Dir()); err != nil {
		return err
	}

	return nil
}

/**
 * New project options and gopath dir
 */
func NewProject(logger *Logger, cwd string) *ProjectImpl {
	var err error

	if "" == cwd {
		panic("Cwd is undefined")
	}

	if cwd, err = filepath.Abs(cwd); err != nil {
		panic(err.Error())
	}

	return &ProjectImpl{
		Logger: logger,
		Cwd:    cwd,
	}
}
