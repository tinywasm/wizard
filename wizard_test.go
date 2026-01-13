package wizard

import (
	"testing"

	"github.com/tinywasm/context"
	"github.com/tinywasm/fmt"
)

type MockStep struct {
	label string
	def   string
	onIn  func(input string, ctx *context.Context) (*context.Context, bool, error)
}

func (m *MockStep) Label() string                            { return m.label }
func (m *MockStep) DefaultValue(ctx *context.Context) string { return m.def }
func (m *MockStep) OnInput(input string, ctx *context.Context) (*context.Context, bool, error) {
	return m.onIn(input, ctx)
}

func TestWizardFlow(t *testing.T) {
	completed := false
	onComplete := func() { completed = true }

	step1 := &MockStep{
		label: "Project Name",
		def:   "my-project",
		onIn: func(input string, ctx *context.Context) (*context.Context, bool, error) {
			c, err := context.WithValue(ctx, "name", input)
			return c, true, err
		},
	}

	step2 := &MockStep{
		label: "Location",
		def:   "./",
		onIn: func(input string, ctx *context.Context) (*context.Context, bool, error) {
			c, err := context.WithValue(ctx, "path", input)
			return c, true, err
		},
	}

	w := New(onComplete, step1, step2)

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
	if w.Name() != "DONE" {
		t.Errorf("expected Name DONE, got %s", w.Name())
	}
	if w.WaitingForUser() {
		t.Error("expected WaitingForUser to be false after completion")
	}

	// Verify context values
	if w.ctx.Value("name") != "test-app" {
		t.Errorf("expected context name test-app, got %s", w.ctx.Value("name"))
	}
	if w.ctx.Value("path") != "/tmp/test" {
		t.Errorf("expected context path /tmp/test, got %s", w.ctx.Value("path"))
	}
}

func TestWizardErrorFlow(t *testing.T) {
	step1 := &MockStep{
		label: "Hard Error Step",
		onIn: func(input string, ctx *context.Context) (*context.Context, bool, error) {
			return ctx, false, fmt.Err("critical failure")
		},
	}

	step2 := &MockStep{
		label: "Soft Error Step",
		onIn: func(input string, ctx *context.Context) (*context.Context, bool, error) {
			return ctx, true, fmt.Err("minor warning")
		},
	}

	step3 := &MockStep{
		label: "Success Step",
		onIn: func(input string, ctx *context.Context) (*context.Context, bool, error) {
			return ctx, true, nil
		},
	}

	w := New(nil, step1, step2, step3)

	// 1. Test Hard Error (continueFlow = false)
	w.Change("trigger crash")
	if w.Label() != "Hard Error Step" {
		t.Errorf("expected to stay on Hard Error Step, got %s", w.Label())
	}
	if w.currentStepIdx != 0 {
		t.Errorf("expected index 0, got %d", w.currentStepIdx)
	}

	// For step 1 to pass, we manually "fix" it for the next Change
	step1.onIn = func(input string, ctx *context.Context) (*context.Context, bool, error) {
		return ctx, true, nil
	}
	w.Change("ok")
	if w.Label() != "Soft Error Step" {
		t.Errorf("expected to move to Soft Error Step, got %s", w.Label())
	}

	// 2. Test Soft Error (continueFlow = true)
	w.Change("trigger warning")
	if w.Label() != "Success Step" {
		t.Errorf("expected to move to Success Step despite warning, got %s", w.Label())
	}
	if w.currentStepIdx != 2 {
		t.Errorf("expected index 2, got %d", w.currentStepIdx)
	}
}
