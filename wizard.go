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

	onComplete func()
}

// New creates a generic wizard with the given steps.
func New(onComplete func(), steps ...*Step) *Wizard {
	iSteps := make([]orchestratorStep, len(steps))
	for i, s := range steps {
		iSteps[i] = s
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
			w.onComplete()
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
