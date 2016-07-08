package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func Test_Runner_RunWithoutName(t *testing.T) {
	runner := &Runner{}
	if err := runner.Run(); err == nil || err.Error() != "Name is undefined" {
		t.Errorf("Must fail if Name is undefined")
	}
}

func Test_Runner_Run(t *testing.T) {
	cwd, err := os.Getwd()

	runner := &Runner{
		Name: "pwd",
		Out:  bytes.NewBuffer([]byte{}),
	}
	err = runner.Run()
	if err != nil {
		t.Error(err.Error())
		return
	}

	runner.Wait()

	if strings.Trim(runner.Out.(*bytes.Buffer).String(), " \n\r") != cwd {
		t.Error("Wrong output or command error")
		return
	}
}

func Test_Runner_RunWithEnv(t *testing.T) {
	runner := &Runner{
		Name: "env",
		Env:  []string{"SOME_SILLY_ENV=foo"},
		Out:  bytes.NewBuffer([]byte{}),
	}

	err := runner.Run()
	if err != nil {
		t.Error(err.Error())
		return
	}

	runner.Wait()

	if !strings.Contains(runner.Out.(*bytes.Buffer).String(), "GOPATH=/") {
		t.Error("System environment variables are not populated")
		return
	}

	if !strings.Contains(runner.Out.(*bytes.Buffer).String(), "SOME_SILLY_ENV=foo") {
		t.Error("Custom Env not appended yet")
		return
	}

	runner = &Runner{
		Name: "pwd",
		Out:  bytes.NewBuffer([]byte{}),
	}

	err = runner.Run()
	if err != nil {
		t.Error(err.Error())
		return
	}

	runner.Wait()
	if cwd, _ := os.Getwd(); strings.Trim(runner.Out.(*bytes.Buffer).String(), "\n\r ") != cwd {
		t.Error("Dir does not use cwd")
		return
	}
}

func Test_Runner_RunOverrideEnv(t *testing.T) {
	runner := &Runner{
		Name: "env",
		Env:  []string{"GOPATH=/foo/bar"},
		Out:  bytes.NewBuffer([]byte{}),
	}

	err := runner.Run()
	if err != nil {
		t.Error(err.Error())
		return
	}

	runner.Wait()

	if !strings.Contains(runner.Out.(*bytes.Buffer).String(), "GOPATH=/foo/bar") {
		t.Error("Variable is not overridden")
		return
	}
}

func Test_Runner_RunOverrideDir(t *testing.T) {
	runner := &Runner{
		Name: "pwd",
		Dir:  "/",
		Out:  bytes.NewBuffer([]byte{}),
	}

	if runner.GetDir() != "/" {
		t.Error("GetDir return wrong")
		return
	}

	err := runner.Run()
	if err != nil {
		t.Error(err.Error())
		return
	}

	runner.Wait()

	if strings.Trim(runner.Out.(*bytes.Buffer).String(), "\n\r ") != "/" {
		t.Error("Dir does not use specified dir")
		return
	}
}
