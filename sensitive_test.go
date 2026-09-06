package wizard

import (
	"strings"
	"testing"

	stdFmt "fmt"

	"webtyp.com/context"
)

func TestStepIsSensitiveDefaultFalse(t *testing.T) {
	s := Step{}
	if s.IsSensitive() {
		t.Fatalf("zero-value Step IsSensitive() = true, want false")
	}
}

func TestStepIsSensitiveTrue(t *testing.T) {
	s := Step{Sensitive: true}
	if !s.IsSensitive() {
		t.Fatalf("Step{Sensitive: true} IsSensitive() = false, want true")
	}
}

func TestWizardSensitiveReflectsCurrentStep(t *testing.T) {
	step1 := &Step{
		LabelText: "Step 1",
		Sensitive: false,
		OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil },
	}
	step2 := &Step{
		LabelText: "Step 2",
		Sensitive: true,
		OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil },
	}
	mod := &MockModule{name: "Test", steps: []*Step{step1, step2}}
	w := New(nil, mod)

	if w.Sensitive() {
		t.Fatalf("wizard on step 1: Sensitive() = true, want false")
	}
	w.Change("value1")
	if !w.Sensitive() {
		t.Fatalf("wizard after advancing to step 2: Sensitive() = false, want true")
	}
}

func TestWizardSensitiveFalseAfterCompletion(t *testing.T) {
	step1 := &Step{
		LabelText: "Only",
		Sensitive: true,
		OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil },
	}
	mod := &MockModule{name: "Test", steps: []*Step{step1}}
	w := New(nil, mod)

	if !w.Sensitive() {
		t.Fatalf("wizard on sensitive step: Sensitive() = false, want true")
	}
	w.Change("secret")
	if w.Sensitive() {
		t.Fatalf("wizard after completion: Sensitive() = true, want false")
	}
}

func TestChangeMasksSensitiveStepInLog(t *testing.T) {
	step1 := &Step{
		LabelText: "Secret",
		Sensitive: true,
		OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil },
	}
	mod := &MockModule{name: "Test", steps: []*Step{step1}}
	w := New(nil, mod)

	var captured []string
	w.SetLog(func(msgs ...any) {
		// Wizard logs single string "✓ Label: value", capture joined representation
		captured = append(captured, stdFmt.Sprint(msgs...))
	})

	w.Change("super-secret-value")

	for _, msg := range captured {
		if strings.Contains(msg, "super-secret-value") {
			t.Fatalf("log must NOT contain raw sensitive value, got %q", msg)
		}
	}
	found := false
	for _, msg := range captured {
		if strings.Contains(msg, "••••••••") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected masked value '••••••••' in log, got %v", captured)
	}
}

func TestChangeLogsPlainStepNormally(t *testing.T) {
	step1 := &Step{
		LabelText: "Normal",
		Sensitive: false,
		OnInputFn: func(input string, ctx *context.Context) (bool, error) { return true, nil },
	}
	mod := &MockModule{name: "Test", steps: []*Step{step1}}
	w := New(nil, mod)

	var captured []string
	w.SetLog(func(msgs ...any) {
		captured = append(captured, stdFmt.Sprint(msgs...))
	})

	w.Change("plain-value")

	found := false
	for _, msg := range captured {
		if strings.Contains(msg, "plain-value") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected plain value 'plain-value' in log, got %v", captured)
	}
}
