package wizard

import (
	"testing"

	"github.com/tinywasm/context"
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
