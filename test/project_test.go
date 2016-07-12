package test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/reekoheek/gopas/util"
)

const (
	TEST_PROJECT_CWD = ".tmp/test"
)

func Test_Project_Dependencies(t *testing.T) {
	test_project_SetUp()
	defer test_project_TearDown()

	project, err := (&util.ProjectImpl{
		Cwd: TEST_PROJECT_CWD,
	}).Construct(nil)

	if err != nil {
		t.Error(err.Error())
		return
	}

	ioutil.WriteFile(
		filepath.Join(TEST_PROJECT_CWD, "gopasfile"),
		[]byte("github.com/reekoheek/foo =\ngithub.com/reekoheek/bar ="),
		0644)

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
