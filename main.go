package main

import "fmt"
import "os"
import "os/exec"
import "io"
import "io/ioutil"
import "strings"
import "path/filepath"
import "bufio"
import "time"
import "runtime"

type Dependency struct {
	Name, Version string
}

type Options struct {
	GOPASFILE  string
	WATCHES    []string
	IGNORES    []string
	EXTENSIONS []string
	MODIFIED   time.Time
}

type Runner struct {
	file    string
	command *exec.Cmd
}

func (r *Runner) Run() (*exec.Cmd, error) {
	if r.command == nil || r.Exited() {
		err := r.runBin()
		// time.Sleep(250 * time.Millisecond)
		return r.command, err
	} else {
		return r.command, nil
	}

}

func (r *Runner) Exited() bool {
	return r.command != nil && r.command.ProcessState != nil && r.command.ProcessState.Exited()
}

func (r *Runner) runBin() error {
	fmt.Printf(">> Running \"go run %s\"\n", r.file)

	r.command = exec.Command("go", "run", r.file)
	env := []string{fmt.Sprintf("GOPATH=%s", cwd()+"/.gopath")}
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "GOPATH=") {
			env = append(env, v)
		}
	}
	r.command.Env = env

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

type ScanCallback func(path string)

var data Options

func readOutPipe(out io.ReadCloser) {
	b := make([]byte, 1)
	for {
		ch, err := out.Read(b)
		if ch > 0 {
			os.Stdout.Write(b)
		} else if err == io.EOF {
			break
		}
	}
}

func readErrPipe(err io.ReadCloser) {
	b := make([]byte, 1)
	for {
		ch, e := err.Read(b)
		if ch > 0 {
			os.Stdout.Write(b)
		} else if e == io.EOF {
			break
		}
	}
}

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

func cwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, ">> Getwd Error: %s\n", err.Error())
		os.Exit(1)
	}
	return cwd
}

func isIgnorable(path string) bool {
	ignorable := false
	for _, ignore := range data.IGNORES {
		if matched, _ := filepath.Match(ignore, path); matched {
			ignorable = true
			break
		}
	}
	return ignorable
}

func isAcceptable(path string) bool {
	for _, ext := range data.EXTENSIONS {
		if strings.HasSuffix(path, "."+ext) {
			return true
		}
	}

	return false
}

func scanChanges(cb ScanCallback) {
	for {
		for _, dir := range data.WATCHES {
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if isIgnorable(path) {
					return filepath.SkipDir
				}

				if isAcceptable(path) && info.ModTime().After(data.MODIFIED) {
					data.MODIFIED = time.Now()
					cb(path)
				}
				return nil
			})
		}

		time.Sleep(1000 * time.Millisecond)
	}
}

func actionList() {
	depCount := 0
	dependencies := readDependencies(data.GOPASFILE)
	for _, dep := range dependencies {
		fmt.Printf("%s %s\n", dep.Name, dep.Version)
		depCount++
	}
	fmt.Printf("dependencies(%d)\n", depCount)
}

func actionInstall() {
	dependencies := readDependencies(data.GOPASFILE)
	for _, dep := range dependencies {
		cmd := exec.Command("go", "get", dep.Name)
		env := []string{fmt.Sprintf("GOPATH=%s", cwd()+"/.gopath")}
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

func actionHelp() {
	fmt.Println("Usage: gopas <action> [<args...>]")
	fmt.Println("")
	fmt.Println("Actions:")
	fmt.Println("  list     List all dependencies")
	fmt.Println("  install  Install dependencies")
	fmt.Println("  run      Run go code")
	fmt.Println("  help     Show help")
}

func actionRun() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gopas run <file>")
		os.Exit(1)
	}

	runner := &Runner{
		file: os.Args[2],
	}

	fmt.Println(">> Watches    ", strings.Join(data.WATCHES, ", "))
	fmt.Println(">> Extensions ", strings.Join(data.EXTENSIONS, ", "))
	fmt.Println(">> Ignores    ", strings.Join(data.IGNORES, ", "))

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

func bootstrap() {
	data = Options{
		GOPASFILE:  "gopasfile",
		WATCHES:    []string{"."},
		IGNORES:    []string{".git", ".gopath"},
		EXTENSIONS: []string{"go"},
		MODIFIED:   time.Now(),
	}

	if _, err := os.Stat(".gopath"); os.IsNotExist(err) {
		if os.MkdirAll(".gopath/src", 0755) != nil {
			fmt.Fprintf(os.Stderr, ">> Bootstrap Error: %s\n", err.Error())
			os.Exit(1)
		}
	}

	cwd := cwd()
	os.Symlink(cwd, ".gopath/src/"+filepath.Base(cwd))
}

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
		}
	}
	actionHelp()
}
