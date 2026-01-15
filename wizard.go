package wizard

import (
	"github.com/tinywasm/context"
)

// orchestratorStep is the internal interface used by the Wizard.
// Both the provided Step struct and any custom implementation can satisfy this.
type orchestratorStep interface {
	Label() string
	DefaultValue(ctx *context.Context) string
	OnInput(input string, ctx *context.Context) (bool, error)
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

// New creates a wizard from items that provide steps (must implement GetSteps() []*Step).
func New(onComplete func(ctx *context.Context), items ...any) *Wizard {
	var iSteps []orchestratorStep

	for _, item := range items {
		if getter, ok := item.(interface{ GetSteps() []*Step }); ok {
			for _, s := range getter.GetSteps() {
				iSteps = append(iSteps, s)
			}
		}
	}

	w := &Wizard{
		log:            func(...any) {},
		ctx:            context.Background(),
		steps:          iSteps,
		currentStepIdx: 0,
		onComplete:     onComplete,
		stepMessage:    "",
	}
	w.initCurrentStep()
	return w
}

func (w *Wizard) initCurrentStep() {
	if w.currentStepIdx >= len(w.steps) {
		w.waitingForUser = false
		if w.onComplete != nil {
			w.onComplete(w.ctx)
			w.onComplete = nil
		}
		return
	}

	step := w.steps[w.currentStepIdx]
	w.label = step.Label()
	w.currentValue = step.DefaultValue(w.ctx)
	w.waitingForUser = true
}
