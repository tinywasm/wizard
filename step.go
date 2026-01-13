package wizard

import "github.com/tinywasm/context"

// Step represents a single interaction or execution unit in the wizard.
// It is designed to be used as a literal for easy instantiation.
type Step struct {
	LabelText string
	DefaultFn func(ctx *context.Context) string
	OnInputFn func(input string, ctx *context.Context) (continueFlow bool, err error)
}

// Label returns the prompt text for the UI.
func (s *Step) Label() string {
	return s.LabelText
}

// DefaultValue returns a suggestion based on the current context.
func (s *Step) DefaultValue(ctx *context.Context) string {
	if s.DefaultFn == nil {
		return ""
	}
	return s.DefaultFn(ctx)
}

// OnInput executes the step logic using a mutable context.
func (s *Step) OnInput(input string, ctx *context.Context) (bool, error) {
	if s.OnInputFn == nil {
		return true, nil
	}
	return s.OnInputFn(input, ctx)
}
