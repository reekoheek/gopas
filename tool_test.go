package main

/**
 * imports
 */
import (
	"bytes"
	"strings"
	"testing"
)

type test_tool_ProjectMock struct {
	isBootstrapped  bool
	isCleaned       bool
	isBuilt         bool
	isRan           bool
	isTested        bool
	willReturnError bool
}

func (p *test_tool_ProjectMock) Bootstrap() error {
	p.isBootstrapped = true
	return nil
}

func (p *test_tool_ProjectMock) Dependencies() []Dependency {
	return []Dependency{
		Dependency{"github.com/reekoheek/foo", ""},
		Dependency{"github.com/reekoheek/bar", ""},
	}
}

func (p *test_tool_ProjectMock) Clean() error {
	p.isCleaned = true
	return nil
}

func (p *test_tool_ProjectMock) Install(dependency Dependency) error {
	return nil
}

func (p *test_tool_ProjectMock) Build() error {
	p.isBuilt = true
	return nil
}

func (p *test_tool_ProjectMock) Run() error {
	p.isRan = true
	return nil
}

func (p *test_tool_ProjectMock) Test() error {
	p.isTested = true
	return nil
}

func test_tool_New() *Tool {
	project := &test_tool_ProjectMock{
		isBootstrapped: false,
	}

	tool := &Tool{
		Project: project,
		Out:     bytes.NewBuffer([]byte{}),
		Err:     bytes.NewBuffer([]byte{}),
	}

	tool.Bootstrap()
	return tool
}

func test_tool_AssertContains(t *testing.T, str string, s string) bool {
	if !strings.Contains(str, s) {
		t.Error("String does not contain " + s)
		return false
	}
	return true
}

func Test_Tool_BootstrapWithoutProject(t *testing.T) {
	tool := &Tool{}
	if err := tool.Bootstrap(); err == nil || err.Error() != "Project is undefined" {
		t.Error("Bootstrap failed")
	}
}

func Test_Tool_Bootstrap(t *testing.T) {
	tool := test_tool_New()
	if !tool.Project.(*test_tool_ProjectMock).isBootstrapped {
		t.Error("Project is not bootstrapped yet")
		return
	}
}

func Test_Tool_DoHelp(t *testing.T) {
	tool := test_tool_New()
	tool.DoHelp(nil)

	out := tool.Out.(*bytes.Buffer).String()

	test_tool_AssertContains(t, out, "  help")
	test_tool_AssertContains(t, out, "  build")
	test_tool_AssertContains(t, out, "  install")
	test_tool_AssertContains(t, out, "  list")
	test_tool_AssertContains(t, out, "  run")
	test_tool_AssertContains(t, out, "  clean")
}

func Test_Tool_DoList(t *testing.T) {
	tool := test_tool_New()
	tool.DoList(nil)

	out := tool.Out.(*bytes.Buffer).String()

	test_tool_AssertContains(t, out, "github.com/reekoheek/foo")
	test_tool_AssertContains(t, out, "github.com/reekoheek/bar")
	test_tool_AssertContains(t, out, "dependencies(2)")
}

func Test_Tool_DoClean(t *testing.T) {
	tool := test_tool_New()
	tool.DoClean(nil)
	if !tool.Project.(*test_tool_ProjectMock).isCleaned {
		t.Error("Project not clean yet")
		return
	}

	out := tool.Out.(*bytes.Buffer).String()
	test_tool_AssertContains(t, out, "Cleaning")
}

func Test_Tool_DoBuild(t *testing.T) {
	tool := test_tool_New()
	tool.DoBuild(nil)
	if !tool.Project.(*test_tool_ProjectMock).isBuilt {
		t.Error("Project not build yet")
		return
	}

	out := tool.Out.(*bytes.Buffer).String()
	test_tool_AssertContains(t, out, "Building")
}

func Test_Tool_DoInstall(t *testing.T) {
	tool := test_tool_New()
	tool.DoInstall(nil)

	out := tool.Out.(*bytes.Buffer).String()
	test_tool_AssertContains(t, out, "github.com/reekoheek/foo@ => ok")
	test_tool_AssertContains(t, out, "github.com/reekoheek/bar@ => ok")
}

func Test_Tool_DoRun(t *testing.T) {
	tool := test_tool_New()
	tool.DoRun(nil)
	if !tool.Project.(*test_tool_ProjectMock).isRan {
		t.Error("Project not run yet")
		return
	}

	out := tool.Out.(*bytes.Buffer).String()
	test_tool_AssertContains(t, out, "Running")
}

func Test_Tool_DoTest(t *testing.T) {
	tool := test_tool_New()
	tool.DoTest(nil)
	if !tool.Project.(*test_tool_ProjectMock).isTested {
		t.Error("Project not test yet")
		return
	}

	out := tool.Out.(*bytes.Buffer).String()
	test_tool_AssertContains(t, out, "Testing")
}
