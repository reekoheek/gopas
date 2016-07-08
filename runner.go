package main

import (
	"errors"
	"fmt"
	"io"
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
	Command *exec.Cmd
}

/**
 * Run command
 *
 * @return {*exec.Cmd}
 * @return {error}
 */
func (r *Runner) Run() error {
	if r.Command == nil || r.IsExited() {
		if r.Name == "" {
			return errors.New("Name is undefined")
		}

		if r.Out == nil {
			r.Out = os.Stdout
		}

		if r.Err == nil {
			r.Err = os.Stderr
		}

		//log.Println("[", r.Name, r.Args, "]")
		r.Command = exec.Command(r.Name, r.Args...)
		r.Command.Env = r.GetEnv()
		r.Command.Dir = r.GetDir()

		stdout, err := r.Command.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := r.Command.StderrPipe()
		if err != nil {
			return err
		}

		err = r.Command.Start()
		if err != nil {
			return err
		}

		go io.Copy(r.Out, stdout)
		go io.Copy(r.Err, stderr)
		go r.Command.Wait()
	}

	return nil
}

func (r *Runner) Wait() error {
	if r.IsExited() {
		return nil
	}

	return r.Command.Wait()
}

/**
 * Check wether runner command already exit
 *
 * @return {bool}
 */
func (r *Runner) IsExited() bool {
	return r.Command != nil && r.Command.ProcessState != nil && r.Command.ProcessState.Exited()
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
		use := true
		for _, prefix := range prefixes {
			if strings.HasPrefix(v, prefix) {
				use = false
				break
			}
		}
		if use {
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
	if r.Command != nil && r.Command.Process != nil {
		done := make(chan error)
		go func() {
			r.Command.Wait()
			close(done)
		}()

		//Trying a "soft" kill first
		if runtime.GOOS == "windows" {
			if err := r.Command.Process.Kill(); err != nil {
				return err
			}
		} else if err := r.Command.Process.Signal(os.Interrupt); err != nil {
			return err
		}

		//Wait for our process to die before we return or hard kill after 3 sec
		select {
		case <-time.After(3 * time.Second):
			if err := r.Command.Process.Kill(); err != nil {
				fmt.Fprintln(os.Stderr, ">> Kill Error: ", err)
				os.Exit(1)
			}
		case <-done:
		}
		r.Command = nil
	}

	return nil
}
