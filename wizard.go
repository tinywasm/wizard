package wizard

import (
	"github.com/tinywasm/context"
	"github.com/tinywasm/fmt"
)

// orchestratorStep is the internal interface used by the Wizard.
// Both the provided Step struct and any custom implementation can satisfy this.
type orchestratorStep interface {
	Label() string
	DefaultValue(ctx *context.Context) string
	OnInput(input string, ctx *context.Context) (bool, error)
}

// Module represents a pluggable component that provides a sequence of steps.
type Module interface {
	Name() string
	GetSteps() []any // Should return []orchestratorStep compatible items
}

// Wizard orchestrates the execution of multiple Steps.
type Wizard struct {
	log func(message ...any)

	// Orchestration state
	ctx            *context.Context
	steps          []orchestratorStep
	currentStepIdx int

	// TUI state
	stepMessage    string
	label          string
	currentValue   string
	waitingForUser bool

	onComplete func(ctx *context.Context)
}

// New creates a wizard from modules that provide steps.
func New(onComplete func(ctx *context.Context), modules ...Module) *Wizard {
	var iSteps []orchestratorStep
	for _, mod := range modules {
		for _, s := range mod.GetSteps() {
			if step, ok := s.(orchestratorStep); ok {
				iSteps = append(iSteps, step)
			}
		}
	}

	w := &Wizard{
		log:            func(...any) {},
		ctx:            context.Background(),
		steps:          iSteps,
		currentStepIdx: 0,
		onComplete:     onComplete,
		stepMessage:    "WIZARD",
	}
	w.initCurrentStep()
	return w
}

func (w *Wizard) initCurrentStep() {
	if w.currentStepIdx >= len(w.steps) {
		// All steps done
		w.waitingForUser = false
		w.label = "Complete"
		w.stepMessage = "DONE"
		w.currentValue = "Wizard completed successfully."

		if w.onComplete != nil {
			w.onComplete(w.ctx)
			w.onComplete = nil
		}
		return
	}

	step := w.steps[w.currentStepIdx]
	w.label = step.Label()
	w.currentValue = step.DefaultValue(w.ctx)
	w.stepMessage = "STEP " + fmt.Convert(w.label).PathBase().String()
	w.waitingForUser = true
}
