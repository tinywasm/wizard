package wizard

import (
	"testing"

	"github.com/tinywasm/context"
	"github.com/tinywasm/fmt"
)

type MockModule struct {
	name  string
	steps []*Step
}

func (m *MockModule) Name() string { return m.name }
func (m *MockModule) GetSteps() []*Step {
	return m.steps
}

func TestWizardFlow(t *testing.T) {
	completed := false
	onComplete := func(ctx *context.Context) { completed = true }

	step1 := &Step{
		LabelText: "Project Name",
		DefaultFn: func(ctx *context.Context) string { return "my-project" },
		OnInputFn: func(input string, ctx *context.Context) (bool, error) {
			err := ctx.Set("name", input)
			return true, err
		},
	}

	step2 := &Step{
		LabelText: "Location",
		DefaultFn: func(ctx *context.Context) string { return "./" },
		OnInputFn: func(input string, ctx *context.Context) (bool, error) {
			err := ctx.Set("path", input)
			return true, err
		},
	}

	mod := &MockModule{name: "Test", steps: []*Step{step1, step2}}
	w := New(onComplete, mod)

	// Verification 1: Initial state
	if w.Label() != "Project Name" {
		t.Errorf("expected Label Project Name, got %s", w.Label())
	}
	if w.Value() != "my-project" {
		t.Errorf("expected Value my-project, got %s", w.Value())
	}
	if !w.WaitingForUser() {
		t.Error("expected WaitingForUser to be true")
	}

	// Verification 2: Transition to step 2
	w.Change("test-app")
	if w.Label() != "Location" {
		t.Errorf("expected Label Location, got %s", w.Label())
	}
	if w.Value() != "./" {
		t.Errorf("expected Value ./, got %s", w.Value())
	}

	// Verification 3: Completion
	w.Change("/tmp/test")
	if !completed {
		t.Error("expected onComplete to have been called")
	}
	if w.WaitingForUser() {
		t.Error("expected WaitingForUser to be false after completion")
	}

	// Verify context values (directly from mutable ctx)
	if w.ctx.Value("name") != "test-app" {
		t.Errorf("expected context name test-app, got %s", w.ctx.Value("name"))
	}
	if w.ctx.Value("path") != "/tmp/test" {
		t.Errorf("expected context path /tmp/test, got %s", w.ctx.Value("path"))
	}
}

func TestWizardErrorFlow(t *testing.T) {
	step1 := &Step{
		LabelText: "Hard Error Step",
		OnInputFn: func(input string, ctx *context.Context) (bool, error) {
			return false, fmt.Err("critical failure")
		},
	}

	step2 := &Step{
		LabelText: "Success Step",
		OnInputFn: func(input string, ctx *context.Context) (bool, error) {
			return true, nil
		},
	}

	mod := &MockModule{name: "ErrorTest", steps: []*Step{step1, step2}}
	w := New(nil, mod)

	// 1. Test Error = stay on step
	w.Change("trigger crash")
	if w.Label() != "Hard Error Step" {
		t.Errorf("expected to stay on Hard Error Step, got %s", w.Label())
	}
	if w.currentStepIdx != 0 {
		t.Errorf("expected index 0, got %d", w.currentStepIdx)
	}

	// 2. Fix step1 logic and advance
	step1.OnInputFn = func(input string, ctx *context.Context) (bool, error) {
		return true, nil
	}
	w.Change("ok")
	if w.Label() != "Success Step" {
		t.Errorf("expected to move to Success Step, got %s", w.Label())
	}
	if w.currentStepIdx != 1 {
		t.Errorf("expected index 1, got %d", w.currentStepIdx)
	}
}
