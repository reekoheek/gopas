package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

/**
 * Runner type
 */
type Runner struct {
	Name    string
	Args    []string
	Env     []string
	Dir     string
	Out     io.Writer
	Err     io.Writer
	command *exec.Cmd
}

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
	if r.Name == "" {
		return errors.New("Name is undefined")
	}

	if r.Out == nil {
		r.Out = os.Stdout
	}

	if r.Err == nil {
		r.Err = os.Stderr
	}

	log.Println("[", r.Name, r.Args, "]")
	r.command = exec.Command(r.Name, r.Args...)
	r.command.Env = r.GetEnv()
	r.command.Dir = r.GetDir()

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

	go io.Copy(r.Out, stdout)
	go io.Copy(r.Err, stderr)
	go r.command.Wait()

	return nil
}

func (r *Runner) GetEnv() []string {
	prefixes := []string{}
	for _, v := range r.Env {
		splitted := strings.Split(v, "=")
		prefixes = append(prefixes, splitted[0]+"=")
	}

	env := []string{}
	env = append(env, r.Env...)

	for _, v := range os.Environ() {
		found := false
		for _, prefix := range prefixes {
			if !strings.HasPrefix(v, prefix) {
				found = true
				break
			}
		}
		if found {
			env = append(env, v)
		}
	}
	return env
}

func (r *Runner) GetDir() string {
	if r.Dir == "" {
		cwd, _ := os.Getwd()
		return cwd
	}

	return r.Dir
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
