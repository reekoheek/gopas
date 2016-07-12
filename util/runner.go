package util

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
	command *exec.Cmd
}

/**
 * Run command
 *
 * @return {*exec.Cmd}
 * @return {error}
 */
func (r *Runner) Run() error {
	if r.command == nil || r.IsExited() {
		var (
			stdout io.ReadCloser
			stderr io.ReadCloser
			err    error
		)

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
		r.command = exec.Command(r.Name, r.Args...)
		r.command.Env = r.GetEnv()
		r.command.Dir = r.GetDir()

		if stdout, err = r.command.StdoutPipe(); err != nil {
			return err
		}

		if stderr, err = r.command.StderrPipe(); err != nil {
			return err
		}

		if err = r.command.Start(); err != nil {
			return err
		}

		go io.Copy(r.Out, stdout)
		go io.Copy(r.Err, stderr)
		go r.command.Wait()
	}

	return nil
}

func (r *Runner) Wait() error {
	return r.command.Wait()
}

/**
 * Check wether runner command already exit
 *
 * @return {bool}
 */
func (r *Runner) IsExited() bool {
	return r.command != nil && r.command.ProcessState != nil && r.command.ProcessState.Exited()
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
	if r.command != nil && r.command.Process != nil {
		done := make(chan error)
		go func() {
			if r.command != nil {
				r.command.Wait()
			}
			close(done)
		}()

		//log.Println("[RUNNER] soft killing ...", os.Args)
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
