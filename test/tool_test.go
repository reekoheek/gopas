package test

/**
 * imports
 */
import (
	"bytes"
	"strings"
	"testing"

	"github.com/reekoheek/gopas/util"
)

type test_tool_ProjectMock struct {
	util.ProjectImpl
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

func (p *test_tool_ProjectMock) Dependencies() []util.Dependency {
	return []util.Dependency{
		util.Dependency{"github.com/reekoheek/foo", ""},
		util.Dependency{"github.com/reekoheek/bar", ""},
	}
}

func (p *test_tool_ProjectMock) Clean() error {
	p.isCleaned = true
	return nil
}

func (p *test_tool_ProjectMock) Install(dependency util.Dependency) error {
	return nil
}

func (p *test_tool_ProjectMock) Name() string {
	return "foo"
}

func (p *test_tool_ProjectMock) Build() error {
	p.isBuilt = true
	return nil
}

func (p *test_tool_ProjectMock) Run(args ...string) error {
	p.isRan = true
	return nil
}

func (p *test_tool_ProjectMock) Test(cover bool, packages ...string) error {
	p.isTested = true
	return nil
}

func test_tool_New() *util.Tool {
	project := &test_tool_ProjectMock{
		isBootstrapped: false,
	}

	tool := &util.Tool{
		Project: project,
	}

	logger := (&util.Logger{
		Out: bytes.NewBuffer([]byte{}),
		Err: bytes.NewBuffer([]byte{}),
	}).Construct()
	tool.Construct(logger)
	return tool
}

func test_tool_AssertContains(t *testing.T, str string, s string) bool {
	if !strings.Contains(str, s) {
		t.Error("String does not contain %s", s)
		return false
	}
	return true
}

func Test_Tool_BootstrapWithoutProject(t *testing.T) {
	tool := &util.Tool{}
	if _, err := tool.Construct(nil); err == nil || err.Error() != "Project is undefined" {
		t.Error("Construct failed")
	}
}

func Test_Tool_DoList(t *testing.T) {
	tool := test_tool_New()
	tool.DoList(nil)

	out := tool.Out.(*bytes.Buffer).String()

	test_tool_AssertContains(t, out, "github.com/reekoheek/foo")
	test_tool_AssertContains(t, out, "github.com/reekoheek/bar")
	test_tool_AssertContains(t, out, "Dependencies foo (2)")
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
	test_tool_AssertContains(t, out, "github.com/reekoheek/foo")
	test_tool_AssertContains(t, out, "github.com/reekoheek/bar")
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
