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

// Handler Interface (compatible with app's TUI expectations)

func (w *Wizard) Name() string         { return w.stepMessage }
func (w *Wizard) Label() string        { return w.label }
func (w *Wizard) Value() string        { return w.currentValue }
func (w *Wizard) WaitingForUser() bool { return w.waitingForUser }

func (w *Wizard) Change(newValue string) {
	if w.currentStepIdx >= len(w.steps) {
		return
	}

	step := w.steps[w.currentStepIdx]
	continueFlow, err := step.OnInput(newValue, w.ctx)

	if err != nil {
		w.log("Error: " + err.Error())
		// If error occurs, we stay on the current step unless we want specific behaviors.
		// For now, any error blocks progress.
		return
	}

	if !continueFlow {
		// Step logic decided not to proceed yet (e.g. waiting for more input)
		return
	}

	// Move to next step
	w.currentStepIdx++
	w.initCurrentStep()
}

// Loggable implementation
func (w *Wizard) SetLog(f func(message ...any)) { w.log = f }

// StreamingLoggable implementation
func (w *Wizard) AlwaysShowAllLogs() bool { return true }

// Cancelable implementation
func (w *Wizard) Cancel() {
	w.waitingForUser = false
	w.log("Wizard cancelled")
}
