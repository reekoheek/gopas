package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const (
	TEST_PROJECT_CWD = ".tmp/test"
)

func Test_Project_BootstrapWithoutCwd(t *testing.T) {
	test_project_SetUp()
	defer test_project_TearDown()

	project := &ProjectImpl{}
	if err := project.Bootstrap(); err == nil || err.Error() != "Cwd is undefined" {
		t.Error("Bootstrap failed")
	}
}

func Test_Project_Bootstrap(t *testing.T) {
	test_project_SetUp()
	defer test_project_TearDown()

	var (
		stat os.FileInfo
		err  error
	)

	project := &ProjectImpl{
		Cwd: TEST_PROJECT_CWD,
	}

	if err = project.Bootstrap(); err != nil {
		t.Error(err.Error())
		return
	}

	stat, err = os.Stat(TEST_PROJECT_CWD + "/.gopath")
	if err != nil {
		t.Error(err.Error())
		return
	}

	if !stat.IsDir() {
		t.Error(".gopath is not directory")
		return
	}

	stat, err = os.Stat(TEST_PROJECT_CWD + "/.gopath/src/test")
	if err != nil {
		t.Error(err.Error())
		return
	}

	if !stat.IsDir() {
		t.Error(".gopath/src/<project> is not directory")
		return
	}

	stat, err = os.Stat(TEST_PROJECT_CWD + "/.gopath/src/foo")
	if err != nil {
		t.Error(err.Error())
		return
	}

	if !stat.IsDir() {
		t.Error(".gopath/src/<vendor> is not directory")
		return
	}
}

func Test_Project_Dependencies(t *testing.T) {
	test_project_SetUp()
	defer test_project_TearDown()

	project := &ProjectImpl{
		Cwd: TEST_PROJECT_CWD,
	}

	ioutil.WriteFile(
		filepath.Join(TEST_PROJECT_CWD, "gopasfile"),
		[]byte("github.com/reekoheek/foo =\ngithub.com/reekoheek/bar ="),
		0644)

	project.Bootstrap()
	dependencies := project.Dependencies()
	if 2 != len(dependencies) {
		t.Error("Dependencies length not matched")
		return
	}

	if dependencies[0].Name != "github.com/reekoheek/foo" {
		t.Error("Dependency data not matched")
	}
}

func test_project_SetUp() {
	os.RemoveAll(TEST_PROJECT_CWD)
	os.MkdirAll(filepath.Join(TEST_PROJECT_CWD, "vendor/foo"), 0755)
}

func test_project_TearDown() {
	os.RemoveAll(TEST_PROJECT_CWD)
}
