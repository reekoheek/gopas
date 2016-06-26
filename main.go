package main

/**
 * Imports
 */
import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

/**
 * Dependency type
 */
type Dependency struct {
	Name, Version string
}

/**
 * Options type
 */
type Options struct {
	GopasFile  string
	Watches    []string
	Ignores    []string
	Extensions []string
}

/**
 * Runner type
 */
type Runner struct {
	file    string
	command *exec.Cmd
}

/**
 * ScanCallback type
 */
type ScanCallback func(path string)

/**
 * Global variables
 */
var (
	options      Options
	modifiedTime time.Time = time.Now()
	cwd          string    = getCwd()
)

/**
 * Run command
 *
 * @return {*exec.Cmd}
 * @return {error}
 */
func (r *Runner) Run() (*exec.Cmd, error) {
	if r.command == nil || r.IsExited() {
		err := r.runBin()
		// time.Sleep(250 * time.Millisecond)
		return r.command, err
	} else {
		return r.command, nil
	}

}

/**
 * Check wether runner command already exit
 *
 * @return {bool}
 */
func (r *Runner) IsExited() bool {
	return r.command != nil && r.command.ProcessState != nil && r.command.ProcessState.Exited()
}

/**
 * Run bin
 *
 * @return {error}
 */
func (r *Runner) runBin() error {
	// cwd := cwd()
	fmt.Printf(">> Running \"go run %s\"\n", r.file)

	r.command = exec.Command("go", "run", r.file)
	// r.command = exec.Command("go", "env")
	env := []string{fmt.Sprintf("GOPATH=%s", cwd+"/.gopath")}
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "GOPATH=") {
			env = append(env, v)
		}
	}
	r.command.Env = env
	r.command.Dir = cwd + "/.gopath/src/" + filepath.Base(cwd)

	stdout, err := r.command.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := r.command.StderrPipe()
	if err != nil {
		return err
	}

	err = r.command.Start()
	if err != nil {
		return err
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	go r.command.Wait()

	return nil
}

/**
 * Kill runner command
 *
 * @return {error}
 */
func (r *Runner) Kill() error {
	if r.command != nil && r.command.Process != nil {
		done := make(chan error)
		go func() {
			r.command.Wait()
			close(done)
		}()

		//Trying a "soft" kill first
		if runtime.GOOS == "windows" {
			if err := r.command.Process.Kill(); err != nil {
				return err
			}
		} else if err := r.command.Process.Signal(os.Interrupt); err != nil {
			return err
		}

		//Wait for our process to die before we return or hard kill after 3 sec
		select {
		case <-time.After(3 * time.Second):
			if err := r.command.Process.Kill(); err != nil {
				fmt.Fprintln(os.Stderr, ">> Kill Error: ", err)
				os.Exit(1)
			}
		case <-done:
		}
		r.command = nil
	}

	return nil
}

/**
 * Read dependencies listed on filename
 *
 * @param {string} filename
 * @return {[]Dependency}
 */
func readDependencies(filename string) []Dependency {
	dependencies := []Dependency{}

	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return dependencies
	}

	fileLines := strings.Split(string(fileBytes), "\n")
	for _, line := range fileLines {
		if line != "" {
			token := strings.Split(line, "=")

			dependencies = append(dependencies, Dependency{
				Name:    strings.Trim(token[0], " \t"),
				Version: strings.Trim(token[1], " \t"),
			})
		}
	}

	return dependencies
}

/**
 * Get current working directory
 *
 * @return {string}
 */
func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, ">> Getwd Error: %s\n", err.Error())
		os.Exit(1)
	}
	return cwd
}

/**
 * Check whether path is ignorable
 *
 * @param {string} path
 */
func isIgnorable(path string) bool {
	ignorable := false
	for _, ignore := range options.Ignores {
		if matched, _ := filepath.Match(ignore, path); matched {
			ignorable = true
			break
		}
	}
	return ignorable
}

/**
 * Check whether path having acceptable extension
 *
 * @param {string} path
 */
func isAcceptable(path string) bool {
	for _, ext := range options.Extensions {
		if strings.HasSuffix(path, "."+ext) {
			return true
		}
	}

	return false
}

/**
 * Scan changes and invoke callback on file changes
 *
 * @param  {ScanCallback}   cb ScanCallback
 */
func scanChanges(cb ScanCallback) {
	for {
		for _, dir := range options.Watches {
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if isIgnorable(path) {
					return filepath.SkipDir
				}

				if isAcceptable(path) && info.ModTime().After(modifiedTime) {
					modifiedTime = time.Now()
					cb(path)
				}
				return nil
			})
		}

		time.Sleep(1000 * time.Millisecond)
	}
}

/**
 * List action
 */
func actionList() {
	depCount := 0
	dependencies := readDependencies(options.GopasFile)
	for _, dep := range dependencies {
		fmt.Printf("%s %s\n", dep.Name, dep.Version)
		depCount++
	}
	fmt.Printf("dependencies(%d)\n", depCount)
}

/**
 * Install action
 */
func actionInstall() {
	dependencies := readDependencies(options.GopasFile)
	for _, dep := range dependencies {
		cmd := exec.Command("go", "get", dep.Name)
		env := []string{fmt.Sprintf("GOPATH=%s", cwd+"/.gopath")}
		for _, v := range os.Environ() {
			if !strings.HasPrefix(v, "GOPATH=") {
				env = append(env, v)
			}
		}
		cmd.Env = env
		cmd.Start()
		if err := cmd.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, ">> Install Error: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf(">> Installing \"go get %s\"\n", dep.Name)
	}
}

/**
 * Help action
 */
func actionHelp() {
	fmt.Println("Gopas is a tool to build Go outside GOPATH")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("")
	fmt.Println("  gopas <action> [<args...>]")
	fmt.Println("")
	fmt.Println("The actions are:")
	fmt.Println("")
	fmt.Println("  build    compile packages and dependencies")
	fmt.Println("  help     show help")
	fmt.Println("  install  compile and install packages and dependencies")
	fmt.Println("  list     list dependencies")
	fmt.Println("  run      compile and run Go program")
	fmt.Println("  test     test packages")
	fmt.Println("")
}

/**
 * Run action
 */
func actionRun() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: gopas run <file>")
		os.Exit(1)
	}

	file := os.Args[2]

	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "File \"%s\" is not exists\n", file)
		os.Exit(1)
	}

	runner := &Runner{
		file: file,
	}

	fmt.Println(">> Watches    ", strings.Join(options.Watches, ", "))
	fmt.Println(">> Extensions ", strings.Join(options.Extensions, ", "))
	fmt.Println(">> Ignores    ", strings.Join(options.Ignores, ", "))

	runner.Run()
	go scanChanges(func(path string) {
		runner.Kill()
		runner.Run()
	})

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "rs" {
			runner.Kill()
			runner.Run()
		}
	}
}

/**
 * Test action
 */
func actionTest() {
	// cwd := cwd()
	cmd := exec.Command("go", "test")
	cmd.Dir = cwd + "/.gopath/src/" + filepath.Base(cwd)
	env := []string{fmt.Sprintf("GOPATH=%s", cwd+"/.gopath")}
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "GOPATH=") {
			env = append(env, v)
		}
	}
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()
}

/**
 * Build action
 */
func actionBuild() {
	// cwd := cwd()
	cmd := exec.Command("go", "build")
	cmd.Dir = cwd + "/.gopath/src/" + filepath.Base(cwd)
	env := []string{fmt.Sprintf("GOPATH=%s", cwd+"/.gopath")}
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "GOPATH=") {
			env = append(env, v)
		}
	}
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()
}

/**
 * Bootstrap project options and gopath dir
 */
func bootstrap() {
	options = Options{
		GopasFile:  "gopasfile",
		Watches:    []string{"."},
		Ignores:    []string{".git", ".gopath"},
		Extensions: []string{"go"},
	}

	if _, err := os.Stat(".gopath"); os.IsNotExist(err) {
		if os.MkdirAll(".gopath/src", 0755) != nil {
			fmt.Fprintf(os.Stderr, ">> Bootstrap Error: %s\n", err.Error())
			os.Exit(1)
		}
	}

	// cwd := cwd()
	os.Symlink(cwd, cwd+"/.gopath/src/"+filepath.Base(cwd))

	// FIXME workaround, cannot read module on vendor dir
	if files, err := ioutil.ReadDir("vendor"); err == nil {
		for _, file := range files {
			name := file.Name()
			os.Symlink(cwd+"/vendor/"+name, cwd+"/.gopath/src/"+name)
		}
	}
}

/**
 * main function
 */
func main() {
	bootstrap()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			actionList()
			return
		case "install":
			actionInstall()
			return
		case "run":
			actionRun()
			return
		case "build":
			actionBuild()
			return
		case "test":
			actionTest()
			return
		}
	}
	actionHelp()
}
